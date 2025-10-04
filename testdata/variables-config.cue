// Additional config with variables and shared targets
package config

variables: {
	app_name: "myapp"
	env: "production"
}

targets: [
	{
		name: "shared-config"
		type: "file"
		config: {
			path: "testdata/example.conf"
			format: "ini"
			content: {
				global: {
					app_name: variables.app_name
					environment: variables.env
				}
			}
		}
	}
]
