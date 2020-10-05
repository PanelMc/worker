package container

import (
	"context"
	"encoding/json"
	"io"
	"sync"

	docker "github.com/docker/docker/api/types"
)

type imagePullEvent struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

var imagePulls struct {
	sync.Mutex
	pending map[string]chan *imagePullEvent
}

func pullImage(ctx context.Context, container *dockerContainer, image string) (<-chan *imagePullEvent, error) {
	imagePulls.Lock()
	defer imagePulls.Unlock()

	ch := imagePulls.pending[image]
	if ch != nil {
		return ch, nil
	}

	if imagePulls.pending == nil {
		imagePulls.pending = make(map[string]chan *imagePullEvent)
	}
	ch = make(chan *imagePullEvent)
	imagePulls.pending[image] = ch

	err := execImagePull(ctx, ch, container, image)

	return ch, err
}

func execImagePull(ctx context.Context, ch chan *imagePullEvent, container *dockerContainer, image string) error {
	container.Logger().Infof("Pulling image %s...", image)
	r, err := container.client.ImagePull(ctx, image, docker.ImagePullOptions{})
	if err != nil {
		close(ch)
		return err
	}

	go func() {
		defer func() {
			r.Close()

			imagePulls.Lock()
			delete(imagePulls.pending, image)
			imagePulls.Unlock()

			close(ch)
		}()

		d := json.NewDecoder(r)
		var event *imagePullEvent
		for {
			if err := d.Decode(&event); err != nil && err == io.EOF {
				break
			}

			ch <- event
		}

		container.Logger().Infof("Image %s pulled!", image)
	}()

	return nil
}
