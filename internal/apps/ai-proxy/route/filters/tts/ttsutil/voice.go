package ttsutil

import "github.com/sashabaranov/go-openai"

var genderMap = map[openai.SpeechVoice]string{
	// female
	openai.VoiceAlloy:   "female",
	openai.VoiceNova:    "female",
	openai.VoiceShimmer: "female",
	"woman":             "female",
	"female":            "female",

	// male
	openai.VoiceEcho:  "male",
	openai.VoiceFable: "male",
	openai.VoiceOnyx:  "male",
	"man":             "male",
	"male":            "male",
}

// MapVoice maps openai.SpeechVoice to service-provider specific voice.
// maleVoice and femaleVoice are the fixed voice names for male and female respectively.
func MapVoice(voice openai.SpeechVoice, maleVoice, femaleVoice string) string {
	gender := genderMap[voice]

	switch gender {
	case "female":
		return femaleVoice
	case "male":
		return maleVoice
	default:
		// return original if mismatch
		return string(voice)
	}
}

// ContentTypeFromFormat returns the MIME type for the given audio format.
func ContentTypeFromFormat(format string) string {
	switch format {
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "aac":
		return "audio/aac"
	case "opus":
		return "audio/opus"
	default:
		return "audio/mpeg"
	}
}
