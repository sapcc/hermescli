package testing

const ListResponse = `
{ 
  "next": "http://hermes-host:8788/v1/events?limit=2&offset=1",
  "previous": "http://hermes-host:8788/v1/events?limit=2&offset=0",
  "events": [
    { 
      "id": "d3f6695e-8a55-5db1-895c-9f7f0910b7a5",
      "eventTime": "2017-11-01T12:28:58.660965+00:00",
      "action": "create/role_assignment",
      "outcome": "success",
      "initiator": {
        "typeURI": "service/security/account/user",
        "id": "21ff350bc75824262c60adfc58b7fd4a7349120b43a990c2888e6b0b88af6398"
      },
      "target": {
        "typeURI": "service/security/account/user",
        "id": "c4d3626f405b99f395a1c581ed630b2d40be8b9701f95f7b8f5b1e2cf2d72c1b"
      },
      "observer": {
        "typeURI": "service/security",
        "id": "0e8a00bf-e36c-5a51-9418-2d56d59c8887"
      }
    }
  ],
  "total": 2
}
`

const GetResponse = `
{
  "typeURI": "http://schemas.dmtf.org/cloud/audit/1.0/event",
  "id": "7189ce80-6e73-5ad9-bdc5-dcc47f176378",
  "eventTime": "2017-12-18T18:27:32.352893+00:00",
  "action": "create",
  "eventType": "activity",
  "outcome": "success",
  "requestPath": "/v2.0/ports.json",
  "reason": {
    "reasonCode": "201",
    "reasonType": "HTTP"
  },
  "initiator": {
    "typeURI": "service/security/account/user",
    "id": "ba8304b657fb4568addf7116f41b4a16",
    "name": "neutron",
    "domain": "Default",
    "project_id": "ba8304b657fb4568addf7116f41b4a16",
    "host": {
      "address": "127.0.0.1",
       "agent": "python-neutronclient"
    }
  },
  "target": {
    "typeURI": "network/port",
    "id": "7189ce80-6e73-5ad9-bdc5-dcc47f176378",
    "project_id": "ba8304b657fb4568addf7116f41b4a16"
  },
  "observer": {
    "typeURI": "service/network",
    "name": "neutron",
    "id": "7189ce80-6e73-5ad9-bdc5-dcc47f176378"
  }
}
`
