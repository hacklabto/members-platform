package auth

type AuthLevel int

const (
	AuthLevel_LoggedOut AuthLevel = iota
	AuthLevel_Applicant
	AuthLevel_Member
	AuthLevel_Board
	AuthLevel_Operations
)
