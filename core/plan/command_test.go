package plan

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/stretchr/testify/require"
)

func TestCommandMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name            string
		command         Command
		expectedJSON    string
		unmarshalString string
	}{
		// Exec
		{
			name:            "exec command without custom name",
			command:         NewExecShellCommand("echo hello", ExecOptions{CustomName: "echo hello"}),
			expectedJSON:    `{"cmd":"sh -c 'echo hello'","customName":"echo hello"}`,
			unmarshalString: "echo hello",
		},
		{
			name:            "exec command with custom name",
			command:         NewExecShellCommand("echo hello", ExecOptions{CustomName: "Say Hello"}),
			expectedJSON:    `{"cmd":"sh -c 'echo hello'","customName":"Say Hello"}`,
			unmarshalString: "RUN#Say Hello:echo hello",
		},

		// Path
		{
			name:            "path command",
			command:         NewPathCommand("/usr/local/bin"),
			expectedJSON:    `{"path":"/usr/local/bin"}`,
			unmarshalString: "PATH:/usr/local/bin",
		},

		// Copy
		{
			name:            "copy command",
			command:         NewCopyCommand("src.txt", "dst.txt"),
			expectedJSON:    `{"src":"src.txt","dest":"dst.txt"}`,
			unmarshalString: "COPY:src.txt dst.txt",
		},

		// File
		{
			name:            "file command without custom name",
			command:         NewFileCommand("/etc/conf", "config.yaml"),
			expectedJSON:    `{"path":"/etc/conf","name":"config.yaml"}`,
			unmarshalString: "FILE:/etc/conf config.yaml",
		},
		{
			name:            "file command with custom name",
			command:         NewFileCommand("/etc/conf", "config.yaml", FileOptions{CustomName: "Config File"}),
			expectedJSON:    `{"path":"/etc/conf","name":"config.yaml","customName":"Config File"}`,
			unmarshalString: "FILE#Config File:/etc/conf config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshalling to JSON object
			data, err := json.Marshal(tt.command)
			require.NoError(t, err, "failed to marshal command")
			require.Equal(t, string(data), tt.expectedJSON, "marshal result")

			// Test unmarshalling from JSON object
			cmd, err := UnmarshalCommand([]byte(tt.expectedJSON))
			require.NoError(t, err, "failed to unmarshal JSON command")

			// Marshal again to verify it produces the same result
			roundTrip, err := json.Marshal(cmd)
			require.NoError(t, err, "failed to marshal unmarshalled command")
			require.Equal(t, string(roundTrip), tt.expectedJSON, "round-trip JSON result")

			// Test unmarshalling from string format
			if tt.unmarshalString != "" {
				cmd, err = UnmarshalCommand([]byte(tt.unmarshalString))
				require.NoError(t, err, "failed to unmarshal string command")

				// Marshal to JSON to verify it produces the same object
				roundTrip, err = json.Marshal(cmd)
				require.NoError(t, err, "failed to marshal string-unmarshalled command")
				require.Equal(t, string(roundTrip), tt.expectedJSON, "string unmarshal to JSON result")

			}
		})
	}
}

func TestShellCommandPreservesBuildKitArguments(t *testing.T) {
	for _, command := range []string{
		"echo hello",
		`printf '%s\n' "it's quoted" > message.txt`,
		`node -e 'require("fs").writeFileSync("message.txt", "hello")'`,
		"printf '%s' \"$HOME\" && echo `printf nested`\nprintf done",
		"",
	} {
		t.Run(command, func(t *testing.T) {
			generated := NewExecShellCommand(command).(ExecCommand)
			info := llb.ExecInfo{State: llb.Scratch()}
			llb.Shlex(generated.Cmd).SetRunOption(&info)
			args, err := info.State.GetArgs(context.Background())
			require.NoError(t, err)
			require.Equal(t, []string{"sh", "-c", command}, args)
		})
	}
}
