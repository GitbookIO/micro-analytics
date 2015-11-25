package web

type PostData struct {
    Time    string              `json:"time"`
    Event   string              `json:"event"`
    Path    string              `json:"path"`
    Ip      string              `json:"ip"`
    Headers map[string]string   `json:"headers"`
}