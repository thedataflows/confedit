package config

targets: [
	{
		name: "test-ini"
		type: "file"
		config: {
			path: "/tmp/test-hooks-fail.ini"
			format: "ini"
			backup: true
			content: {
				"section1": {
					"key1": "value1"
				}
			}
		}
	}
]

// Test hook with intentional failure
hooks: {
	pre_apply: [
		"""
		#!/bin/bash
		echo "=== PRE-APPLY HOOK (WILL FAIL) ==="
		echo "This hook will fail intentionally"
		echo "Testing error handling..."
		exit 1
		"""
	]

	post_apply: [
		"""
		#!/bin/bash
		echo "=== POST-APPLY HOOK ==="
		echo "This should not be reached due to pre-apply failure"
		"""
	]
}
