package models

type QualityProfilesSearch struct {
	Profiles []Profiles `json:"profiles"`
	Actions  Actions    `json:"actions,omitempty"`
}

type Profiles struct {
	Key                       string         `json:"key"`
	Name                      string         `json:"name"`
	Language                  string         `json:"language"`
	LanguageName              string         `json:"languageName,omitempty"`
	IsInherited               bool           `json:"isInherited,omitempty"`
	IsBuiltIn                 bool           `json:"isBuiltIn,omitempty"`
	ActiveRuleCount           int            `json:"activeRuleCount,omitempty"`
	ActiveDeprecatedRuleCount int            `json:"activeDeprecatedRuleCount,omitempty"`
	IsDefault                 bool           `json:"isDefault"`
	RuleUpdatedAt             string         `json:"ruleUpdatedAt,omitempty"`
	LastUsed                  string         `json:"lastUsed,omitempty"`
	Actions                   ProfileActions `json:"actions,omitempty"`
	ParentKey                 string         `json:"parentKey,omitempty"`
	ParentName                string         `json:"parentName,omitempty"`
	ProjectCount              int            `json:"projectCount,omitempty"`
	UserUpdatedAt             string         `json:"userUpdatedAt,omitempty"`
}

type ProfileActions struct {
	Edit              bool `json:"edit"`
	SetAsDefault      bool `json:"setAsDefault"`
	Copy              bool `json:"copy"`
	Delete            bool `json:"delete"`
	AssociateProjects bool `json:"associateProjects"`
}

type Actions struct {
	Create bool `json:"create"`
}

type SettingsValues struct {
	Setting []Setting `json:"settings"`
}

type Setting struct {
	Key         string           `json:"key"`
	Value       string           `json:"value,omitempty"`
	Inherited   bool             `json:"inherited"`
	Values      []string         `json:"values,omitempty"`
	FieldValues []SettingsValues `json:"fieldValues,omitempty"`
}

type SettingFieldValue struct {
	Boolean string `json:"boolean"`
	Text    string `json:"text"`
}

type QualityGate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	IsBuiltIn bool   `json:"isBuiltIn"`
	Actions   struct {
		Rename            bool `json:"rename"`
		SetAsDefault      bool `json:"setAsDefault"`
		Copy              bool `json:"copy"`
		AssociateProjects bool `json:"associateProjects"`
		Delete            bool `json:"delete"`
		ManageConditions  bool `json:"manageConditions"`
	} `json:"actions"`
}

type QualityActions struct {
	Create bool `json:"create"`
}

type QualityGatesList struct {
	QualityFates []QualityGate  `json:"qualitygates"`
	Default      int            `json:"default"`
	Actions      QualityActions `json:"actions"`
}
