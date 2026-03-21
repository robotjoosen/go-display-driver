package sprite

import (
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v4"
)

var registry = xsync.NewMap[SpriteType, *DecodedSprite]()

type DecodedSprite struct {
	Image    image.Image
	Filepath string
	Modified time.Time
	mu       sync.RWMutex
}

func LoadAll(basePath string) error {
	spriteTypes := []SpriteType{
		SpriteLogo,
	}

	for _, st := range spriteTypes {
		filepath := filepath.Join(basePath, "sprites", string(st)+".png")
		if err := Load(st, filepath); err != nil {
			return err
		}
	}

	return nil
}

func Load(spriteType SpriteType, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	img, err := png.Decode(file)
	if err != nil {
		return err
	}

	registry.Store(spriteType, &DecodedSprite{
		Image:    img,
		Filepath: filepath,
		Modified: stat.ModTime(),
	})

	return nil
}

func Get(spriteType SpriteType) (image.Image, bool) {
	sprite, ok := registry.Load(spriteType)
	if !ok {
		return nil, false
	}

	sprite.mu.RLock()
	defer sprite.mu.RUnlock()

	return sprite.Image, true
}

func StartFileWatcher(interval time.Duration, basePath string) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			registry.Range(func(key SpriteType, value *DecodedSprite) bool {
				if err := value.checkAndReload(); err != nil {
					slog.Warn("failed to reload sprite",
						slog.String("sprite", string(key)),
						slog.String("err", err.Error()),
					)
				}
				return true
			})
		}
	}()
}

func (s *DecodedSprite) checkAndReload() error {
	s.mu.RLock()
	currentModified := s.Modified
	filepath := s.Filepath
	s.mu.RUnlock()

	stat, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if stat.ModTime().After(currentModified) {
		s.mu.Lock()
		defer s.mu.Unlock()

		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			return err
		}

		s.Image = img
		s.Modified = stat.ModTime()

		slog.Info("sprite reloaded",
			slog.String("path", filepath),
		)
	}

	return nil
}
