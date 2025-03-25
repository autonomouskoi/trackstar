import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as slpb from "/m/trackstarstagelinq/pb/stagelinq_pb.js";
import { ControlPanel } from "/tk.js";

const TOPIC_REQUEST = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_REQUEST);

let help = document.createElement('div');
help.innerHTML = `Devices detected by <code>trackstarstagelinq</code> are shown here.`;

class Devices extends ControlPanel {
    constructor() {
        super({ title: 'Detected Devices', help });

        this.refresh();
    }

    refresh() {
        bus.waitForTopic(TOPIC_REQUEST, 5000)
            .then(() => {
                return bus.sendAnd(new buspb.BusMessage({
                    topic: TOPIC_REQUEST,
                    type: slpb.MessageTypeRequest.GET_DEVICES_REQ,
                    message: new slpb.GetDevicesRequest().toBinary(),
                }))
            }).then((reply) => {
                let resp = slpb.GetDevicesResponse.fromBinary(reply.message);
                this.update(resp.devices);
            });
    }

    update(devices: slpb.Device[]) {
        this.innerHTML = '';
        devices.forEach((device) => this.appendChild(new StagelinQDevice(device)));
    }
}
customElements.define('trackstar-stagelinq-devices', Devices, { extends: 'fieldset' });

class StagelinQDevice extends HTMLElement {
    private _device: slpb.Device;

    constructor(device: slpb.Device) {
        super();
        this.attachShadow({ mode: 'open' });
        this.device = device;
    }

    set device(device: slpb.Device) {
        this._device = device;
        let trackData = this._device.services.some((svc) => {
            return svc.name == 'StateMap';
        })
        this.shadowRoot!.innerHTML = `
<style>
fieldset {
    width: fit-content;
    display: grid;
    grid-template-columns: auto auto;
    column-gap: 1rem;
}
</style>
<fieldset>
        <legend>IP: ${this._device.ip}</legend>
        <div>Name</div>
        <div>${this._device.name}</div>
        <div>Software Name</div>
        <div>${this._device.softwareName}</div>
        <div>Software Version</div>
        <div>${this._device.softwareVerison}</div>
        <div>Track Data</div>
        <div>${trackData ? "&#x2705;" : "&#x274C"}</div>
</fieldset>
`;
    }
}
customElements.define('trackstar-stagelinq-device', StagelinQDevice);

export { Devices };