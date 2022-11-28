package usecases

import (
	"errors"

	"github.com/gregszalay/ocpp-charging-station-simulator/stationmessages"
	"github.com/gregszalay/ocpp-messages-go/types/BootNotificationResponse"
	"github.com/gregszalay/ocpp-messages-go/types/StatusNotificationRequest"
)

type Provisioning struct {
	validator_map ValidatorMap
}

func NewProvisioning() Provisioning {
	cold_boot_charging_station := ValidatorList{
		{
			// The Charging Station is powered up
			// The Charging Station sends BootNotificationRequest to the CSMS.
			stationmessages.Create_BootNotificationRequest(),
			// The CSMS returns with BootNotificationResponse with the status Accepted.
			map[string]interface{}{"status": BootNotificationResponse.RegistrationStatusEnumType_1_Accepted},
		},
		{
			// The Charging Station sends StatusNotificationRequest to the CSMS for each Connector.
			stationmessages.Create_StatusNotificationRequest(StatusNotificationRequest.ConnectorStatusEnumType_1_Available),
			nil,
		},
		{
			// Normal operational is resumed.
			// The Charging Station sends HeartbeatRequest to the CSMS.
			stationmessages.Create_HeartbeatRequest(),
			nil,
		},
	}

	p := Provisioning{}
	m := ValidatorMap{
		"B01": cold_boot_charging_station,
	}
	p.validator_map = m
	return p
}

func (p Provisioning) GetValidatorList(usecase string) (ValidatorList, error) {
	if v, ok := p.validator_map[usecase]; ok {
		return v, nil
	}
	return nil, errors.New("Can not find validator list from " + usecase)
}
