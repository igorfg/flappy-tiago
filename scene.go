package main

import (
	"fmt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

type scene struct {
	bg         *sdl.Texture
	bird       *bird
	pipes      *pipes
	razerSound *mix.Chunk
}

func newScene(r *sdl.Renderer) (*scene, error) {
	bg, err := img.LoadTexture(r, "res/img/background.png")
	if err != nil {
		return nil, fmt.Errorf("could not load background image: %v", err)
	}

	b, err := newBird(r)
	if err != nil {
		return nil, err
	}

	ps, err := newPipes(r)
	if err != nil {
		return nil, err
	}

	rs, err := mix.LoadWAV("res/music/razor.wav")
	if err != nil {
		return nil, fmt.Errorf("could not load razor sound: %v", err)
	}

	return &scene{bg: bg, bird: b, pipes: ps, razerSound: rs}, nil
}

func (s *scene) run(events <-chan sdl.Event, r *sdl.Renderer) <-chan error {
	errc := make(chan error)

	go func() {
		defer close(errc)
		tick := time.Tick(10 * time.Millisecond)
		for {
			select {
			case e := <-events:
				if done := s.handleEvent(e); done {
					return
				}
			case <-tick:
				s.update()
				if s.bird.isDead() {
					s.paintGameover(r)
					mix.PauseMusic()
					s.razerSound.Play(-1, 0)
					time.Sleep(3 * time.Second)
					mix.ResumeMusic()
					s.restart()
				}

				if err := s.paint(r); err != nil {
					errc <- err
				}
			}
		}
	}()
	return errc
}

func (s *scene) handleEvent(event sdl.Event) bool {
	switch event.(type) {
	case *sdl.QuitEvent:
		return true
	case *sdl.MouseButtonEvent:
		s.bird.jump()
	case *sdl.MouseMotionEvent, *sdl.WindowEvent, *sdl.TouchFingerEvent:
		// do nothing
	default:
		log.Printf("unknown event %T", event)
	}
	return false

}

func (s *scene) update() {
	s.bird.update()
	s.pipes.update()
	s.pipes.touch(s.bird)
}

func (s *scene) restart() {
	s.bird.restart()
	s.pipes.restart()
}

func (s *scene) paint(r *sdl.Renderer) error {
	r.Clear()

	if err := r.Copy(s.bg, nil, nil); err != nil {
		return fmt.Errorf("could not copy background: %v", err)
	}

	if err := s.bird.paint(r); err != nil {
		return err
	}

	if err := s.pipes.paint(r); err != nil {
		return err
	}

	r.Present()
	return nil
}

func (s *scene) destroy() {
	s.bg.Destroy()
	s.bird.destroy()
	s.pipes.destroy()
}

func (s *scene) paintGameover(r *sdl.Renderer) error {
	gameOver, err := img.LoadTexture(r, "res/img/GAME-OVER.png")
	if err != nil {
		return fmt.Errorf("could not load game over image: %v", err)
	}
	if err := r.Copy(gameOver, nil, nil); err != nil {
		return fmt.Errorf("could not copy game over: %v", err)
	}

	r.Present()
	return nil
}
