package events

type Event struct {
	Name                         string                 `json:"name,omitempty"`
	ID                           string                 `json:"id,omitempty"`
	PreviousIds                  string                 `json:"previousIds,omitempty"`
	PreviousNames                string                 `json:"previousNames,omitempty"`
	Publisher                    string                 `json:"publisher,omitempty"`
	ReplyTo                      string                 `json:"replyTo,omitempty"`
	ResourceID                   string                 `json:"resourceId,omitempty"`
	ResourceType                 string                 `json:"resourceType,omitempty"`
	Transitioning                string                 `json:"transitioning,omitempty"`
	TransitioningInternalMessage string                 `json:"transitioningInternalMessage,omitempty"`
	TransitioningMessage         string                 `json:"transitioningMessage,omitempty"`
	TransitioningProgress        string                 `json:"transitioningProgress,omitempty"`
	Data                         map[string]interface{} `json:"data,omitempty"`
}

type ReplyEvent struct {
	Name        string                 `json:"name"`
	PreviousIds []string               `json:"previousIds"`
	Data        map[string]interface{} `json:"data"`
}

func NewReplyEvent(replyTo string, eventID string) *ReplyEvent {
	return &ReplyEvent{Name: replyTo, PreviousIds: []string{eventID}}
}

/*
{
    "context": {
        "logicName": "demo",
        "logicPath": "physicalhost.activate->(demo)",
        "prettyProcess": "physicalhost.activate",
        "prettyResource": "physicalHost:1",
        "processId": "30",
        "processName": "physicalhost.activate",
        "processUuid": "2649d4f9-5695-4f38-b49b-3e0b257ff325",
        "resouceId": "1",
        "resouceType": "physicalHost",
        "topProcessName": "physicalhost.activate",
        "topResourceId": "1",
        "topResourceType": "physicalHost"
    },
    "data": {
        "driver": "virtualbox",
        "kind": "dockerMachine",
        "name": "test-random-280937",
        "virtualboxMemory": "2048"
		"virtualboxDiskSize":
        "virtualboxBoot2dockerUrl":
        "digitaloceanImage":
        "digitaloceanRegion":
        "digitaloceanSize":
        "digitaloceanAccessToken":
    },
    "id": "190ad7e5-fa1d-4e28-97a2-b9b1bad3f6a8",
    "name": "physicalhost.activate;handler=demo",
    "previousIds": null,
    "previousNames": null,
    "publisher": null,
    "replyTo": "reply.7884953948567153747",
    "resourceId": "1ph1",
    "resourceType": "physicalHost",
    "time": 1419876894816,
    "timeoutMillis": 15000,
    "transitioning": null,
    "transitioningInternalMessage": null,
    "transitioningMessage": null,
    "transitioningProgress": null
}

*/
