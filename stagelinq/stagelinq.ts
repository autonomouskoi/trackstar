import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as stagelinqpb from "/m/trackstarstagelinq/pb/stagelinq_pb.js";

const TOPIC_STAGELINQ_STATE = enumName(stagelinqpb.BusTopics, stagelinqpb.BusTopics.STAGELINQ_STATE);
const TOPIC_STAGELINQ_CONTROL = enumName(stagelinqpb.BusTopics, stagelinqpb.BusTopics.STAGELINQ_CONTROL);

class StagelinQDevice extends HTMLElement {
    private _device: stagelinqpb.Device;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    set device(device: stagelinqpb.Device) {
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

customElements.define('stagelinq-device', StagelinQDevice);

function start(mainContainer: HTMLElement) {
    let button = document.createElement('button');
    button.innerText = 'Capture Fader Threshold';

    let thresholdInput = document.createElement('input') as HTMLInputElement;
    thresholdInput.readOnly = true;
    thresholdInput.value = "0.0";

    mainContainer.style.setProperty('display', 'flex');
    mainContainer.style.setProperty('flex-direction', 'column');

    let thresholdDiv = document.createElement('div') as HTMLDivElement;
    thresholdDiv.style.setProperty('display', 'flex');
    thresholdDiv.style.setProperty('flex-direction', 'row');
    
    thresholdDiv.appendChild(button);
    thresholdDiv.appendChild(thresholdInput);

    let description = document.createElement('fieldset') as HTMLFieldSetElement;
    description.innerHTML = `
<legend>Usage</legend>
<p>To set the threshold:
    <ol>
        <li>Bring all your faders all the way down</li>
        <li>Bring one up to the level you want to display the next song</li>
        <li>Click <em>Capture Fader Threshold</em></li>
    </ol>
</p>
<p>The tool should save the threshold between sessions. In future sessions, however, the tool won't be aware of each fader's current level until you have moved it at least once.</p>
`;

    mainContainer.appendChild(thresholdDiv);
    mainContainer.appendChild(description);

    let devicesDiv = document.createElement('div') as HTMLElement;
    mainContainer.appendChild(devicesDiv);
    let updateDevices = (devices: stagelinqpb.Device[]) => {
        devicesDiv.textContent = '';
        devices.forEach((device) => {
            let sld = document.createElement('stagelinq-device') as StagelinQDevice;
            sld.device = device;
            devicesDiv.appendChild(sld);
            console.log(`appended ${device.ip}`);
        });
        //devicesDiv.innerText = JSON.stringify(devices);
    }

    bus.subscribe(TOPIC_STAGELINQ_STATE, (msg: buspb.BusMessage) => {
        switch (msg.type) {
            case stagelinqpb.MessageType.TYPE_THRESHOLD_UPDATE:
                button.disabled = false;
                let resp = stagelinqpb.ThresholdUpdate.fromBinary(msg.message);
                thresholdInput.value = resp.faderThreshold.toString();
                break;
            case stagelinqpb.MessageType.TYPE_GET_DEVICES_RESPONSE:
                let deviceDiscovered = stagelinqpb.GetDevicesResponse.fromBinary(msg.message);
                updateDevices(deviceDiscovered.devices);
        }
    });

    button.onclick = () => {
        button.disabled = true;
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_STAGELINQ_CONTROL;
        msg.type = stagelinqpb.MessageType.TYPE_CAPTURE_THRESHOLD_REQUEST;
        msg.message = (new stagelinqpb.CaptureThresholdRequest()).toBinary();
        bus.send(msg);
    }


    bus.waitForTopic(TOPIC_STAGELINQ_CONTROL, 5000)
        .then(() => {
            let initialGTRequest = new buspb.BusMessage();
            initialGTRequest.topic = TOPIC_STAGELINQ_CONTROL;
            initialGTRequest.type = stagelinqpb.MessageType.TYPE_GET_THRESHOLD_REQUEST;
            bus.send(initialGTRequest);

            let initialGDRequest = new buspb.BusMessage();
            initialGDRequest.topic = TOPIC_STAGELINQ_CONTROL;
            initialGDRequest.type = stagelinqpb.MessageType.TYPE_GET_DEVICES_REQUEST;
            bus.sendWithReply(initialGDRequest, (resp: buspb.BusMessage) => {
                if (resp.type !== stagelinqpb.MessageType.TYPE_GET_DEVICES_RESPONSE) {
                    return;
                }
                let gdr = stagelinqpb.GetDevicesResponse.fromBinary(resp.message);
                updateDevices(gdr.devices);
            });
        });
}

export { start };