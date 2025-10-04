// Example CUE configuration file for setting CheckSpace in INI options section
package config

targets: [
	{
		name: "example-ini-config"
		type: "file"
		config: {
			path:   "testdata/example.conf"
			format: "ini"
			options: {
				use_spacing: false // Add spaces before and after separator for new keys (default: true)
			}
			content: {
				direct: "value1"
				"": {
					rootkey: "rootvalue"
					a: {value: null, commented: "; "}
				}
				options: {
					CheckSpace: {value: "x", commented: "; "}
					deletedkey: {deleted: true}
					new: "new_value"
				}
			}
		}
	},
]
