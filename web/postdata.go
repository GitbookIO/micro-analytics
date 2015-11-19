package web

type PostData struct {
    Time    string              `json:"time"`
    Type    string              `json:"type"`
    Path    string              `json:"path"`
    Ip      string              `json:"ip"`
    Headers map[string]string   `json:"headers"`
}