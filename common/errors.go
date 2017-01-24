package common

import "fmt"

// public enum Type
const (
		// id have to be kept synchronised with the server side.
		OK = 0
		WRONG_CONFIGURATION = 1
		AUTHENTICATION_FAILED = 2
		TOO_OLD_CLIENT = 3
		SESSION_DISABLED = 4
		RENDERER_NOT_AVAILABLE = 5
		MISSING_RENDER = 6
		MISSING_SCENE = 7
		NOOUTPUTFILE = 8
		DOWNLOAD_FILE = 9
		CAN_NOT_CREATE_DIRECTORY = 10
		NETWORK_ISSUE = 11
		RENDERER_CRASHED = 12
		RENDERER_OUT_OF_VIDEO_MEMORY = 13

		RENDERER_KILLED = 14
		RENDERER_MISSING_LIBRARIES = 15
		FAILED_TO_EXECUTE = 16
		OS_NOT_SUPPORTED = 17
		CPU_NOT_SUPPORTED = 18
		GPU_NOT_SUPPORTED = 19
		RENDERER_KILLED_BY_USER = 20
		RENDERER_OUT_OF_MEMORY = 21
		RENDERER_KILLED_BY_SERVER = 22
		
		// internal error handling
		UNKNOWN = 99
		NO_SPACE_LEFT_ON_DEVICE = 100
		ERROR_BAD_RESPONSE = 101
)
	
// public enum ServerCode
const (
		SERVER_OK = 0
		SERVER_UNKNOWN = 999
		
		CONFIGURATION_ERROR_NO_CLIENT_VERSION_GIVEN = 100
		CONFIGURATION_ERROR_CLIENT_TOO_OLD = 101
		CONFIGURATION_ERROR_AUTH_FAILED = 102
		CONFIGURATION_ERROR_WEB_SESSION_EXPIRED = 103
		CONFIGURATION_ERROR_MISSING_PARAMETER = 104
		
		JOB_REQUEST_NOJOB = 200
		JOB_REQUEST_ERROR_NO_RENDERING_RIGHT = 201
		JOB_REQUEST_ERROR_DEAD_SESSION = 202
		JOB_REQUEST_ERROR_SESSION_DISABLED = 203
		JOB_REQUEST_ERROR_INTERNAL_ERROR = 204
		JOB_REQUEST_ERROR_RENDERER_NOT_AVAILABLE = 205
		JOB_REQUEST_SERVER_IN_MAINTENANCE = 206
		JOB_REQUEST_SERVER_OVERLOADED = 207
		
		JOB_VALIDATION_ERROR_MISSING_PARAMETER = 300
		JOB_VALIDATION_ERROR_BROKEN_MACHINE = 301 // in GPU the generated frame is black
		JOB_VALIDATION_ERROR_FRAME_IS_NOT_IMAGE = 302
		JOB_VALIDATION_ERROR_UPLOAD_FAILED = 303
		JOB_VALIDATION_ERROR_SESSION_DISABLED = 304 // missing heartbeat or broken machine
		
		KEEPMEALIVE_STOP_RENDERING = 400
		
		// internal error handling
		SERVER_ERROR_NO_ROOT = 2
		SERVER_ERROR_BAD_RESPONSE = 3
		SERVER_ERROR_REQUEST_FAILED = 5
)
	
func ServerCodeToError(sc int) int {
	switch (sc) {
	case SERVER_OK:
		return OK
	case SERVER_UNKNOWN:
		return UNKNOWN
	case CONFIGURATION_ERROR_CLIENT_TOO_OLD:
		return TOO_OLD_CLIENT
	case CONFIGURATION_ERROR_AUTH_FAILED:
		return AUTHENTICATION_FAILED
		
	case CONFIGURATION_ERROR_NO_CLIENT_VERSION_GIVEN, CONFIGURATION_ERROR_WEB_SESSION_EXPIRED:
		return WRONG_CONFIGURATION
		
	case JOB_REQUEST_ERROR_SESSION_DISABLED, JOB_VALIDATION_ERROR_SESSION_DISABLED:
		return SESSION_DISABLED
		
	case JOB_REQUEST_ERROR_RENDERER_NOT_AVAILABLE:
		return RENDERER_NOT_AVAILABLE
	
	default:
		return UNKNOWN
	}
}
	
func ErrorAsString(in int) string {
	switch (in) {
	case ERROR_BAD_RESPONSE:
		return "Bad answer from server. It's a server side error, wait a bit and retry later."
	case NETWORK_ISSUE:
		return "Could not connect to the server, please check if you have connectivity issue"
	case TOO_OLD_CLIENT:
		return "This client is too old, you need to update it"
	case AUTHENTICATION_FAILED:
		return "Failed to authenticate, please check your login and password"
	case DOWNLOAD_FILE:
		return "Error while downloading project files. Will try another project in a few minutes."
	case NOOUTPUTFILE:
		return "Renderer has generated no output file, possibly a wrong project configuration or you are missing required libraries. Will try another project in a few minutes."
	case RENDERER_CRASHED:
		return "Renderer has crashed. It's usually due to a bad project or not enough memory. There is nothing you can do about it. Will try another project in a few minutes."
	case RENDERER_OUT_OF_VIDEO_MEMORY:
		return "Renderer has crashed, due to not enough video memory (vram). There is nothing you can do about it. Will try another project in a few minutes."
	case RENDERER_OUT_OF_MEMORY:
		return "No more memory available. There is nothing you can do about it. Will try another project in a few minutes."
	case GPU_NOT_SUPPORTED:
		return "Rendering have failed because your GPU is not supported"
	case RENDERER_MISSING_LIBRARIES:
		return "Failed to launch renderer. Please check if you have necessary libraries installed and if you have enough free space in your working directory."
	case RENDERER_KILLED:
		return "The renderer stopped because either you asked to stop or the server did (usually for a render time too high)."
	case RENDERER_KILLED_BY_USER:
		return "The renderer stopped because you've blocked its project."
	case RENDERER_KILLED_BY_SERVER:
		return "The renderer stopped because it's been killed by the server. Usually because the project will take too much time or it's been paused."
	case SESSION_DISABLED:
		return "The server has disabled your session. Your client may have generated a broken frame (GPU not compatible, not enough RAM/VRAM, etc)."
	case RENDERER_NOT_AVAILABLE:
		return "No renderer are available on the server for your machine."
	case OS_NOT_SUPPORTED:
		return "Operating System not supported."
	case CPU_NOT_SUPPORTED:
		return "CPU not supported."
	case NO_SPACE_LEFT_ON_DEVICE:
		return "No space left on hard disk"
	default:
		return fmt.Sprintf("SheepIt Error Code %d", in);
	}
}