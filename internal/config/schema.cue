package config

import "list"

// Metadata structure for all targets
#Metadata: {
	description?: string
	tags?: [...string]
	priority?: int & >=0 & <=100
	...
}

// INI value types (explicit separation for better validation)
#INISimpleValue: string | null

#INICommentedValue: {
	value: string | null
	commented: "; " | "# "
}

#INIDeletedValue: {
	deleted: true
}

#INIValue: #INISimpleValue | #INICommentedValue | #INIDeletedValue

// INI-specific configuration options with defaults
#INIOptions: {
	// Add spaces before and after separator for new keys
	use_spacing: *true | bool

	// Characters to recognize as comment prefixes
	comment_chars: *"#;" | string

	// Key-value delimiter character (must be single character)
	delimiter: *"=" | string
}

// INI content structure
#INIContent: {
	[key=string]: #INIValue | {
		[subkey=string]: #INIValue
	}
}

// File configuration schema
#FileConfig: {
	path: string & !=""
	format: "ini" | "yaml" | "toml" | "json" | "xml"
	owner?: string
	group?: string
	mode?: string
	backup: *true | bool

	if format == "ini" {
		options?: #INIOptions
		content: #INIContent
	}

	if format != "ini" {
		content: {...}
	}
}

// Dconf configuration schema
#DconfConfig: {
	user?: string
	schema: string & !=""
	settings: {
		[key=string]: string | bool | int | float | [...string]
	}
}

// Systemd configuration schema
#SystemdConfig: {
	unit: string & !=""
	section: string & !=""
	properties: {
		[key=string]: string | bool | int | float
	}
	reload: *false | bool
}

// Sed configuration schema
#SedConfig: {
	path: string & !=""
	commands: [...string] & list.MinItems(1)
	backup: *true | bool
	options?: {
		[key=string]: string | bool
	}
}

// Shell script validation
#ShellScript: string & !=""

// Target type definitions (explicit types for better validation)
#FileTarget: {
	name!: string
	type: "file"
	metadata?: #Metadata
	config: #FileConfig
}

#DconfTarget: {
	name!: string
	type: "dconf"
	metadata?: #Metadata
	config: #DconfConfig
}

#SystemdTarget: {
	name!: string
	type: "systemd"
	metadata?: #Metadata
	config: #SystemdConfig
}

#SedTarget: {
	name!: string
	type: "sed"
	metadata?: #Metadata
	config: #SedConfig
}

// Union of all target types
#ConfigTarget: #FileTarget | #DconfTarget | #SystemdTarget | #SedTarget

// Top-level system configuration
#SystemConfig: {
	targets: [...#ConfigTarget]
	variables?: {
		[key=string]: _
	}
	hooks?: {
		pre_apply?: [...#ShellScript]
		post_apply?: [...#ShellScript]
	}
}
