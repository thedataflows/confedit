package config

targets: [
	{
		name: "test-ini"
		type: "file"
		config: {
			path: "/tmp/test-hooks.ini"
			format: "ini"
			backup: true
			content: {
				"section1": {
					"key1": "value1"
					"key2": "value2"
				}
			}
		}
	}
]

// Global hooks demonstrating full shell script capabilities
hooks: {
	pre_apply: [
		"""
		#!/bin/bash
		echo "=== PRE-APPLY HOOK START ==="
		echo "Current timestamp: $(date)"
		echo "Environment check:"
		if [ -d "/tmp" ]; then
			echo "✓ /tmp directory exists"
		else
			echo "✗ /tmp directory missing"
			exit 1
		fi

		# Multi-line script with conditionals and loops
		echo "Checking for test files:"
		for file in /tmp/test-hooks.*; do
			if [ -f "$file" ]; then
				echo "  Found: $file"
			fi
		done

		echo "Creating pre-apply marker..."
		echo "pre-apply-$(date +%s)" > /tmp/pre-apply-marker.txt
		echo "=== PRE-APPLY HOOK END ==="
		"""
	]

	post_apply: [
		"""
		#!/bin/bash
		echo "=== POST-APPLY HOOK START ==="
		echo "Configuration applied at: $(date)"

		# Verify the file was created
		if [ -f "/tmp/test-hooks.ini" ]; then
			echo "✓ Configuration file created successfully"
			echo "Content preview:"
			head -5 /tmp/test-hooks.ini | sed 's/^/  /'
		else
			echo "✗ Configuration file was not created"
			exit 1
		fi

		# Clean up markers
		if [ -f "/tmp/pre-apply-marker.txt" ]; then
			echo "✓ Found pre-apply marker, removing..."
			rm -f /tmp/pre-apply-marker.txt
		fi

		echo "=== POST-APPLY HOOK END ==="
		"""
	]
}
