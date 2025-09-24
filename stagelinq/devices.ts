import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as slpb from "/m/trackstarstagelinq/pb/stagelinq_pb.js";
import { ControlPanel } from "/tk.js";

const TOPIC_EVENT = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_EVENT);
const TOPIC_REQUEST = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_REQUEST);

let help = document.createElement('div');
help.innerHTML = `Devices detected by <code>trackstarstagelinq</code> are shown here.`;

class Devices extends ControlPanel {
    private _devices: { [key: string]: slpb.Device } = {};

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
                resp.devices.forEach((device) => this._devices[device.token] = device);
                bus.subscribe(TOPIC_EVENT, (msg) => this._handleEvent(msg));
                this._render();
            });
    }

    private _handleEvent(msg: buspb.BusMessage) {
        if (msg.type != slpb.MessageTypeEvent.DEVICE_STATE) {
            return;
        }
        let event = slpb.DeviceStateEvent.fromBinary(msg.message);
        this._devices[event.device.token] = event.device;
        this._render();
    }

    private _render() {
        this.innerHTML = '';
        Object.keys(this._devices).toSorted().forEach((token) =>
            this.appendChild(new StagelinQDevice(this._devices[token]))
        );
    }
}
customElements.define('trackstar-stagelinq-devices', Devices, { extends: 'fieldset' });

class StagelinQDevice extends HTMLFieldSetElement {
    private _device: slpb.Device;

    constructor(device: slpb.Device) {
        super();
        this.style.display = 'inline-block';
        this.device = device;
    }

    set device(device: slpb.Device) {
        this._device = device;
        this.innerHTML = `
<legend>Token: ${this._device.token}</legend>
<div class="grid-2-col">
    <label>Status</label>
    <div>${enumName(slpb.DeviceStatus, this._device.status)}</div>
    <label>Status Detail</label>
    <div>${this._device.statusDetail}</div>
    <label>Name</label>
    <div>${this._device.name}</div>
    <label>IP</label>
    <div>${this._device.ip}</div>
    <label>Software Name</label>
    <div>${this._device.softwareName}</div>
    <label>Software Version</label>
    <div>${this._device.softwareVerison}</div>
</div>
`;
    }
}
customElements.define('trackstar-stagelinq-device', StagelinQDevice, { extends: 'fieldset' });

export { Devices };