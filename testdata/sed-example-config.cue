package config

// Example configuration demonstrating sed target usage
// This shows how to use sed commands to modify text files

variables: {
	config_path: "./testdata/sed-example.conf"
	backup_enabled: true
}

targets: [
	{
		name: "sed-config-example"
		type: "sed"
		config: {
			path: variables.config_path
			commands: [
				// Replace all occurrences of "old_value" with "new_value"
				"s/old_value/new_value/g",
				// Delete lines containing "DEBUG"
				"/DEBUG/d",
				// Add a comment at the beginning
				"1i# Modified by confedit",
				// Replace port numbers
				"s/port=8080/port=9090/g",
			]
			backup: variables.backup_enabled
		}
	},
	{
		name: "simple-sed-example"
		type: "sed"
		config: {
			path: variables.config_path
			commands: [
				// Simple find and replace
				"s/localhost/127.0.0.1/g",
				// Comment out lines starting with "debug"
				"s/^debug/#debug/",
			]
			backup: false
		}
	}
]

// Optional hooks for sed operations
hooks: {
	pre_apply: [
		"echo 'Starting sed operations...'",
		"# You can add validation here"
	]
	post_apply: [
		"echo 'Sed operations completed successfully'",
		"# You can add verification here"
	]
}
