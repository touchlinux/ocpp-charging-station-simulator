package usecases

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-simulator/stationmessages"
	"github.com/gregszalay/ocpp-messages-go/types/AuthorizeResponse"
	"github.com/gregszalay/ocpp-messages-go/types/StatusNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
)

type Transactions struct {
	validator_map ValidatorMap
}

func NewTransactions() Transactions {
	transaction_message_id := uuid.New().String()

	cable_plugin_first := ValidatorList{
		{
			// The EV Driver plugs in the cable at the Charging Station.
			// The Charging Station sends a StatusNotificationRequest to the CSMS to inform it about a Connector that became Occupied.
			stationmessages.Create_StatusNotificationRequest(StatusNotificationRequest.ConnectorStatusEnumType_1_Occupied),
			nil,
		},
		{
			// The Charging Station sends a TransactionEventRequest (eventType = Started) notifying the CSMS about a transaction that has started (even when the driver is not yet known.)
			stationmessages.Create_TransactionEventRequest(TransactionEventRequest.TransactionEventEnumType_1_Started, 0, 22.0, TransactionEventRequest.TriggerReasonEnumType_1_CablePluggedIn, transaction_message_id),
			// The CSMS responds with a TransactionEventResponse, confirming that the TransactionEventRequest was received.
			nil,
		},
		{
			// The EV Driver is authorized by the Charging Station and/or CSMS.
			stationmessages.Create_AuthorizeRequest(),
			map[string]interface{}{"status": AuthorizeResponse.AuthorizationStatusEnumType_1_Accepted},
		},
		{
			// The energy offer starts.
			// The Charging Station sends a TransactionEventRequest (eventType = Updated) with the authorized idToken information to the CSMS to inform about the charging status and which idToken belongs to the transaction.
			stationmessages.Create_TransactionEventRequest(TransactionEventRequest.TransactionEventEnumType_1_Updated, 0.1, 22.0, TransactionEventRequest.TriggerReasonEnumType_1_Authorized, transaction_message_id),
			// The CSMS responds with a TransactionEventResponse to the Charging Station with the IdTokenInfo.status Accepted.
			nil},
		{
			// During the charging process, the Charging Stations continues to send TransactionEventRequest (Updated) messages for transaction-related notifications.
			stationmessages.Create_TransactionEventRequest(TransactionEventRequest.TransactionEventEnumType_1_Updated, 0.1, 22.0, TransactionEventRequest.TriggerReasonEnumType_1_ChargingStateChanged, transaction_message_id),
			nil},
		{
			//
			stationmessages.Create_AuthorizeRequest(),
			nil},
		{
			//
			stationmessages.Create_TransactionEventRequest(TransactionEventRequest.TransactionEventEnumType_1_Updated, 0.1, 22.0, TransactionEventRequest.TriggerReasonEnumType_1_StopAuthorized, transaction_message_id),
			nil},
		{
			stationmessages.Create_StatusNotificationRequest(StatusNotificationRequest.ConnectorStatusEnumType_1_Available),
			nil},
		{
			stationmessages.Create_TransactionEventRequest(TransactionEventRequest.TransactionEventEnumType_1_Ended, 0.1, 22.0, TransactionEventRequest.TriggerReasonEnumType_1_EVCommunicationLost, transaction_message_id),
			nil},
	}

	t := Transactions{}
	m := ValidatorMap{
		"E02": cable_plugin_first,
	}
	t.validator_map = m
	return t
}

func (t Transactions) GetValidatorList(usecase string) (ValidatorList, error) {
	if v, ok := t.validator_map[usecase]; ok {
		return v, nil
	}
	return nil, errors.New("Can not find validator list from " + usecase)
}
