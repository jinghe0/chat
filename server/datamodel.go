package main

/******************************************************************************
 *
 *  Description :
 *
 *    Wire protocol structures
 *
 *****************************************************************************/

import (
	"net/http"
	"strings"
	"time"
)

type JsonDuration time.Duration

func (jd *JsonDuration) UnmarshalJSON(data []byte) (err error) {
	d, err := time.ParseDuration(strings.Trim(string(data), "\""))
	*jd = JsonDuration(d)
	return err
}

type MsgBrowseOpts struct {
	// Load messages/ranges with IDs equal or greater than this
	SinceId int `json:"since,omitempty"`
	// Load messages/ranges with IDs lower than this
	BeforeId int `json:"before,omitempty"`
	// Limit the number of messages loaded
	Limit int `json:"limit,omitempty"`
}

type MsgGetOpts struct {
	IfModifiedSince *time.Time `json:"ims,omitempty"`
	Limit           int        `json:"limit,omitempty"`
}

type MsgGetQuery struct {
	What string `json:"what"`

	// Parameters of "desc" request
	Desc *MsgGetOpts `json:"desc,omitempty"`
	// Parameters of "sub" request
	Sub *MsgGetOpts `json:"sub,omitempty"`
	// Parameters of "data" request
	Data *MsgBrowseOpts `json:"data,omitempty"`
	// Parameters of "del" request
	Del *MsgBrowseOpts `json:"del,omitempty"`
}

// MsgSetSub: payload in set.sub request to update current subscription or invite another user, {sub.what} == "sub"
type MsgSetSub struct {
	// User affected by this request. Default (empty): current user
	User string `json:"user,omitempty"`

	// Access mode change, either Given or Want depending on context
	Mode string `json:"mode,omitempty"`
}

// MsgSetDesc: C2S in set.what == "desc" and sub.init message
type MsgSetDesc struct {
	DefaultAcs *MsgDefaultAcsMode `json:"defacs,omitempty"` // default access mode
	Public     interface{}        `json:"public,omitempty"`
	Private    interface{}        `json:"private,omitempty"` // Per-subscription private data
}

type MsgSetQuery struct {
	// Topic metadata, new topic & new subscriptions only
	Desc *MsgSetDesc `json:"desc,omitempty"`
	// Subscription parameters
	Sub *MsgSetSub `json:"sub,omitempty"`
}

// fndXXX.private is set to this object.
type MsgFindQuery struct {
	// List of tags to query for. Tags of the form "email:jdoe@example.com" or "tel:18005551212"
	Tags []string `json:"tags"`
}

// Either an individual ID or a randge of deleted IDs
type MsgDelQuery struct {
	SeqId int `json:"seq,omitempty"`
	LowId int `json:"low,omitempty"`
	HiId  int `json:"hi,omitempty"`
}

// Client to Server (C2S) messages

// Handshake {hi} message
type MsgClientHi struct {
	// Message Id
	Id string `json:"id,omitempty"`
	// User agent
	UserAgent string `json:"ua,omitempty"`
	// Protocol version, i.e. "0.13"
	Version string `json:"ver,omitempty"`
	// Client's unique device ID
	DeviceID string `json:"dev,omitempty"`
	// ISO 639-1 human language of the connected device
	Lang string `json:"lang,omitempty"`
}

// User creation message {acc}
type MsgClientAcc struct {
	// Message Id
	Id string `json:"id,omitempty"`
	// "newXYZ" to create a new user or UserId to update a user; default: current user
	User string `json:"user,omitempty"`
	// The initial authentication scheme the account can use
	Scheme string `json:"scheme,omitempty"`
	// Shared secret
	Secret []byte `json:"secret,omitempty"`
	// Authenticate session with the newly created account
	Login bool `json:"login"`
	// Indexable tags for user discovery
	Tags []string `json:"tags"`
	// User initialization data when creating a new user, otherwise ignored
	Desc *MsgSetDesc `json:"desc,omitempty"`
}

// Login {login} message
type MsgClientLogin struct {
	// Message Id
	Id string `json:"id,omitempty"`
	// Authentication scheme
	Scheme string `json:"scheme,omitempty"`
	// Shared secret
	Secret []byte `json:"secret"`
}

// Subscription request {sub} message
type MsgClientSub struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`

	// mirrors {set}
	Set *MsgSetQuery `json:"set,omitempty"`

	// mirrors {get}
	Get *MsgGetQuery `json:"get,omitempty"`
}

const (
	constMsgMetaDesc = 1 << iota
	constMsgMetaSub
	constMsgMetaData
	constMsgMetaDel
	constMsgDelTopic
	constMsgDelMsg
	constMsgDelSub
)

func parseMsgClientMeta(params string) int {
	var bits int
	parts := strings.SplitN(params, " ", 8)
	for _, p := range parts {
		switch p {
		case "desc":
			bits |= constMsgMetaDesc
		case "sub":
			bits |= constMsgMetaSub
		case "data":
			bits |= constMsgMetaData
		case "del":
			bits |= constMsgMetaDel
		default:
			// ignore
		}
	}
	return bits
}

func parseMsgClientDel(params string) int {
	var bits int

	switch params {
	case "", "msg":
		return constMsgDelMsg
	case "topic":
		return constMsgDelTopic
	case "sub":
		return constMsgDelSub
	default:
		// ignore
	}
	return bits
}

// Topic default access mode
type MsgDefaultAcsMode struct {
	Auth string `json:"auth,omitempty"`
	Anon string `json:"anon,omitempty"`
}

// Unsubscribe {leave} request message
type MsgClientLeave struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`
	Unsub bool   `json:unsub,omitempty`
}

// MsgClientPub is client's request to publish data to topic subscribers {pub}
type MsgClientPub struct {
	Id      string            `json:"id,omitempty"`
	Topic   string            `json:"topic"`
	NoEcho  bool              `json:"noecho,omitempty"`
	Head    map[string]string `json:"head,omitempty"`
	Content interface{}       `json:"content"`
}

// Query topic state {get}
type MsgClientGet struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`
	MsgGetQuery
}

// Update topic state {set}
type MsgClientSet struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`
	MsgSetQuery
}

// MsgClientDel delete messages or topic
type MsgClientDel struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`
	// What to delete, either "msg" to delete messages (default) or "topic" to delete the topic or "sub"
	// to delete a subscription to topic.
	What string `json:"what"`
	// Delete messages with these IDs (either one by one or a set of ranges)
	DelSeq []MsgDelQuery `json:"delseq,omitempty"`
	// User ID of the subscription to delete
	User string `json:"user,omitempty"`
	// Request to hard-delete messages for all users, if such option is available.
	Hard bool `json:"hard,omitempty"`
}

// MsgClientNote is a client-generated notification for topic subscribers
type MsgClientNote struct {
	// There is no Id -- server will not akn {ping} packets, they are "fire and forget"
	Topic string `json:"topic"`
	// what is being reported: "recv" - message received, "read" - message read, "kp" - typing notification
	What string `json:"what"`
	// Server-issued message ID being reported
	SeqId int `json:"seq,omitempty"`
}

type ClientComMessage struct {
	Hi    *MsgClientHi    `json:"hi"`
	Acc   *MsgClientAcc   `json:"acc"`
	Login *MsgClientLogin `json:"login"`
	Sub   *MsgClientSub   `json:"sub"`
	Leave *MsgClientLeave `json:"leave"`
	Pub   *MsgClientPub   `json:"pub"`
	Get   *MsgClientGet   `json:"get"`
	Set   *MsgClientSet   `json:"set"`
	Del   *MsgClientDel   `json:"del"`
	Note  *MsgClientNote  `json:"note"`

	// from: userid as string
	from      string
	timestamp time.Time
}

/////////////////////////////////////////////////////////////
// Server to client messages

// MsgLastSeenInfo contains info on user's appearance online - when & user agent
type MsgLastSeenInfo struct {
	// Timestamp of user's last appearance online.
	When *time.Time `json:"when,omitempty"`
	// User agent of the device when the user was last online.
	UserAgent string `json:"ua,omitempty"`
}

type MsgAccessMode struct {
	// Access mode requested by the user
	Want string `json:"want,omitempty"`
	// Access mode granted to the user by the admin
	Given string `json:"given,omitempty"`
	// Cumulative access mode want & given
	Mode string `json:"mode,omitempty"`
}

// Topic description, S2C in Meta message
type MsgTopicDesc struct {
	CreatedAt *time.Time `json:"created,omitempty"`
	UpdatedAt *time.Time `json:"updated,omitempty"`
	// When a group topic is created, it's given a temporary name by the client.
	// Then this name changes. Report the original name here.
	TempName   string             `json:"tmpname,omitempty"`
	DefaultAcs *MsgDefaultAcsMode `json:"defacs,omitempty"`
	// Actual access mode
	Acs *MsgAccessMode `json:"acs,omitempty"`
	// Max message ID
	SeqId     int `json:"seq,omitempty"`
	ReadSeqId int `json:"read,omitempty"`
	RecvSeqId int `json:"recv,omitempty"`
	// Id if the last delete operation
	DelId  int         `json:"del,omitempty"`
	Public interface{} `json:"public,omitempty"`
	// Per-subscription private data
	Private interface{} `json:"private,omitempty"`
}

// MsgTopicSub: topic subscription details, sent in Meta message
type MsgTopicSub struct {
	// Fields common to all subscriptions

	// Timestamp when the subscription was last updated
	UpdatedAt *time.Time `json:"updated,omitempty"`
	// Timestamp when the subscription was deleted
	DeletedAt *time.Time `json:"deleted,omitempty"`

	// If the subscriber/topic is online
	Online bool `json:"online,omitempty"`

	// Access mode. Topic admins receive the full info, non-admins receive just the cumulative mode
	// Acs.Mode = want & given. The field is not a pointer because at least one value is always assigned.
	Acs MsgAccessMode `json:"acs"`
	// ID of the message reported by the given user as read
	ReadSeqId int `json:"read,omitempty"`
	// ID of the message reported by the given user as received
	RecvSeqId int `json:"recv,omitempty"`
	// Topic's public data
	Public interface{} `json:"public,omitempty"`
	// User's own private data per topic
	Private interface{} `json:"private,omitempty"`

	// Response to non-'me' topic

	// Uid of the subscribed user
	User string `json:"user,omitempty"`

	// The following sections maks sense only in context of getting
	// user's own subscriptions ('me' topic response)

	// Topic name of this subscription
	Topic string `json:"topic,omitempty"`
	// ID of the last {data} message in a topic
	SeqId int `json:"seq,omitempty"`
	// Id of the latest Delete operation
	DelId int `json:"clear,omitempty"`

	// P2P topics only:

	// Other user's last online timestamp & user agent
	LastSeen *MsgLastSeenInfo `json:"seen,omitempty"`
}

type MsgServerCtrl struct {
	Id     string      `json:"id,omitempty"`
	Topic  string      `json:"topic,omitempty"`
	Params interface{} `json:"params,omitempty"`

	Code      int       `json:"code"`
	Text      string    `json:"text,omitempty"`
	Timestamp time.Time `json:"ts"`
}

/*
// Action announcement: invitation to a join, approval of a request to join, access change,
// subscription gone: topic deleted/unsubscribed.
// Sent as MsgServerData.Content
type MsgAnnounce struct {
	// Topic that user wants to subscribe to or is invited to
	Topic string `json:"topic"`
	// User being subscribed
	User string `json:"user"`
	// Type of this invite - AnnInv, AnnAppr, AnnUpd, AnnDel (defined in store/types/)
	Action string `json:"act"`
	// Current state of the access mode
	Acs *MsgAccessMode `json:"acs,omitempty"`
	// Request made at this authentication level
	AuthLevel string `json:"authlvl,omitempty"`
	// Free-form info passed unchanged from the client
	Info SubInfo `json:"info,omitempty"`
}
*/

type MsgServerData struct {
	Topic string `json:"topic"`
	// ID of the user who originated the message as {pub}, could be empty if sent by the system
	From      string            `json:"from,omitempty"`
	Timestamp time.Time         `json:"ts"`
	DeletedAt *time.Time        `json:"deleted,omitempty"`
	SeqId     int               `json:"seq"`
	Head      map[string]string `json:"head,omitempty"`
	Content   interface{}       `json:"content"`
}

type MsgServerPres struct {
	Topic     string         `json:"topic"`
	Src       string         `json:"src"`
	What      string         `json:"what"`
	UserAgent string         `json:"ua,omitempty"`
	SeqId     int            `json:"seq,omitempty"`
	DelSeq    []MsgDelQuery  `json:"delseq,omitempty"`
	AcsTarget string         `json:"tgt,omitempty"`
	AcsActor  string         `json:"act,omitempty"`
	Acs       *MsgAccessMode `json:"acs,omitempty"`

	// UNroutable params

	// Flag to break the reply loop
	wantReply bool

	// Additional access mode filter when senting to topic's online members
	filter int

	// When sending to 'me', skip sessions subscribed to this topic
	skipTopic string

	// Send to sessions of a single user only
	singleUser string
}

type MsgServerMeta struct {
	Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`

	Timestamp *time.Time `json:"ts,omitempty"`

	// Topic description
	Desc *MsgTopicDesc `json:"desc,omitempty"`
	// Subscriptions as an array of objects
	Sub []MsgTopicSub `json:"sub,omitempty"`
	// List of IDs of deleted messages
	Del []MsgDelQuery `json:"del,omitempty"`
}

// MsgServerInfo is the server-side copy of MsgClientNote with From added
type MsgServerInfo struct {
	Topic string `json:"topic"`
	// ID of the user who originated the message
	From string `json:"from"`
	// what is being reported: "rcpt" - message received, "read" - message read, "kp" - typing notification
	What string `json:"what"`
	// Server-issued message ID being reported
	SeqId int `json:"seq,omitempty"`
}

type ServerComMessage struct {
	Ctrl *MsgServerCtrl `json:"ctrl,omitempty"`
	Data *MsgServerData `json:"data,omitempty"`
	Meta *MsgServerMeta `json:"meta,omitempty"`
	Pres *MsgServerPres `json:"pres,omitempty"`
	Info *MsgServerInfo `json:"info,omitempty"`

	// to: topic
	rcptto string
	// Originating session to send an aknowledgement to. Used only for {data} messages. Could be nil.
	sessFrom *Session
	// MsgServerData has no Id field, copying it here for use in {ctrl} aknowledgements
	id string
	// timestamp for consistency of timestamps in {ctrl} messages
	timestamp time.Time
	// Should the packet be sent to the original sessions? SessionIDs to skip.
	skipSid string
}

// Generators of error messages

func NoErr(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusOK, // 200
		Text:      "ok",
		Topic:     topic,
		Timestamp: ts}}
}

func NoErrCreated(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusCreated, // 201
		Text:      "created",
		Topic:     topic,
		Timestamp: ts}}
}

func NoErrAccepted(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusAccepted, // 202
		Text:      "accepted",
		Topic:     topic,
		Timestamp: ts}}
}

func NoErrEvicted(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusResetContent, // 205
		Text:      "evicted",
		Topic:     topic,
		Timestamp: ts}}
}

func NoErrShutdown(ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Code:      http.StatusResetContent, // 205
		Text:      "server shutdown",
		Timestamp: ts}}
}

// 3xx
func InfoAlreadySubscribed(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotModified, // 304
		Text:      "already subscribed",
		Topic:     topic,
		Timestamp: ts}}
}

func InfoNotJoined(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotModified, // 304
		Text:      "not joined",
		Topic:     topic,
		Timestamp: ts}}
}

func InfoNoAction(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotModified, // 304
		Text:      "no action",
		Topic:     topic,
		Timestamp: ts}}
}

func InfoNotModified(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotModified, // 304
		Text:      "not modified",
		Topic:     topic,
		Timestamp: ts}}
}

// 4xx Errors
func ErrMalformed(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusBadRequest, // 400
		Text:      "malformed",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAuthRequired(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusUnauthorized, // 401
		Text:      "authentication required",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAuthFailed(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusUnauthorized, // 401
		Text:      "authentication failed",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAuthUnknownScheme(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusUnauthorized, // 401
		Text:      "unknown authentication scheme",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrPermissionDenied(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusForbidden, // 403
		Text:      "permission denied",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrTopicNotFound(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotFound,
		Text:      "topic not found", // 404
		Topic:     topic,
		Timestamp: ts}}
}

func ErrUserNotFound(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotFound, // 404
		Text:      "user not found or offline",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAlreadyAuthenticated(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusConflict, // 409
		Text:      "already authenticated",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrDuplicateCredential(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusConflict, // 409
		Text:      "duplicate credential",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAttachFirst(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusConflict, // 409
		Text:      "must attach first",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrAlreadyExists(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusConflict, // 409
		Text:      "already exists",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrCommandOutOfSequence(id, unused string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusConflict, // 409
		Text:      "command out of sequence",
		Timestamp: ts}}
}

func ErrGone(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusGone, // 410
		Text:      "gone",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrPolicy(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusUnprocessableEntity, // 422
		Text:      "policy violation",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrLocked(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusLocked, // 423
		Text:      "locked",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrUnknown(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusInternalServerError, // 500
		Text:      "internal error",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrNotImplemented(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusNotImplemented, // 501
		Text:      "not implemented",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrClusterNodeUnreachable(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusBadGateway, // 502
		Text:      "unreachable",
		Topic:     topic,
		Timestamp: ts}}
}

func ErrVersionNotSupported(id, topic string, ts time.Time) *ServerComMessage {
	return &ServerComMessage{Ctrl: &MsgServerCtrl{
		Id:        id,
		Code:      http.StatusHTTPVersionNotSupported, // 505
		Text:      "version not supported",
		Topic:     topic,
		Timestamp: ts}}
}
