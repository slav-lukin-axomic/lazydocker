package gui

import (
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/i18n"
	"github.com/jesseduffield/lazydocker/pkg/oscommand"
	"github.com/stretchr/testify/assert"
)

// TestAttachToContainer covers the guards that reject attaching before details
// are loaded, when stdin isn't open, or when the container has exited, plus the
// happy path where the docker CLI command is returned without executing it.
func TestAttachToContainer(t *testing.T) {
	t.Parallel()

	tr := i18n.NewTranslationSet(oscommand.NewDummyLog(), "en")

	cases := []struct {
		name      string
		container *domain.Container
		wantErr   string
		wantCmd   bool
	}{
		{
			name:      "details not loaded",
			container: &domain.Container{ID: "abc123", Details: nil},
			wantErr:   tr.WaitingForContainerInfo,
		},
		{
			name:      "stdin not open",
			container: &domain.Container{ID: "abc123", Details: &domain.ContainerDetails{OpenStdin: false}},
			wantErr:   tr.UnattachableContainerError,
		},
		{
			name: "exited container",
			container: &domain.Container{
				ID:      "abc123",
				Status:  domain.StatusExited,
				Details: &domain.ContainerDetails{OpenStdin: true},
			},
			wantErr: tr.CannotAttachStoppedContainerError,
		},
		{
			name: "attachable running container",
			container: &domain.Container{
				ID:      "abc123",
				Status:  domain.StatusRunning,
				Details: &domain.ContainerDetails{OpenStdin: true},
			},
			wantCmd: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gui := &Gui{
				Log:       oscommand.NewDummyLog(),
				OSCommand: oscommand.NewDummyOSCommand(),
				Tr:        tr,
			}

			cmd, err := gui.attachToContainer(tc.container)

			if tc.wantCmd {
				assert.NoError(t, err)
				assert.NotNil(t, cmd)
				assert.Equal(t, []string{"docker", "attach", "--sig-proxy=false", tc.container.ID}, cmd.Args)
				return
			}
			assert.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
			assert.Nil(t, cmd)
		})
	}
}
