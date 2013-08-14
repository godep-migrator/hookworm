package hookworm

// TravisPayload is the struct representation of a JSON payload from Travis
type TravisPayload struct {
	ID         int               `json:"id"`
	Repository *TravisRepository `json:"repository"`
	Number     string            `json:"number"`
	Config     *TravisConfig     `json"config"`
	Status     int               `json:"status"`
	Result     int               `json:"result"`
}

// TravisRepository is how a repository is represented in a TravisPayload
type TravisRepository struct {
}

// TravisConfig is the JSON representation of the '.travis.yml' file
type TravisConfig struct {
}
