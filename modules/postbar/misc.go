package postbar

import ()

//Merry------------------v
const DefaultDeviceID = "1642C31E46901488362D9ADEE56AAF95"

func GenCUID(DeviceID, PhoneIMEI string) string {
	if PhoneIMEI == "" {
		PhoneIMEI = "0"
	}
	var PhoneIMEI_runes = []rune(PhoneIMEI)
	var received = make([]rune, len(PhoneIMEI_runes))
	for i := 0; i < len(PhoneIMEI_runes); i++ {
		received[len(PhoneIMEI_runes)-1-i] = PhoneIMEI_runes[i]
	}
	if DeviceID == "" {
		DeviceID = DefaultDeviceID
	}
	return DefaultDeviceID + "|" + string(received)
}
