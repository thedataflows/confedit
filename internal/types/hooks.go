package types

type Hooks struct {
	PreApply  []string `json:"pre_apply,omitempty"`
	PostApply []string `json:"post_apply,omitempty"`
}
