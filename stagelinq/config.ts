import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as slpb from '/m/trackstarstagelinq/pb/stagelinq_pb.js';
import { Cfg } from './controller.js';
import { UpdatingControlPanel } from '/tk.js';

const TOPIC_REQUEST = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_REQUEST);

let help = document.createElement('div');
help.innerHTML = `
<p>
<em>Fader Threshold</em> tells <code>trackstarstagelinq</code> to not send a
track until the fader for that channel is above a certain level. To set the
threshold:
</p>

<ol>
<li>Make sure <code>trackstarstagelinq</code> is connected to your device.</li>
<li>Bring all faders all the way up</li>
<li>Bring all faders all the way down</li>
<li>Pick any fader; bring it up to the desired level</li>
<li>Click <em>Capture Fader Threshold</em></li>
</ol>

<p>
The threshold value isn't linear; having the fader halfway up does not result
in a value of <code>0.5</code>. Don't set a threshold all the way up; you'll
find the track doesn't always get sent when you expect. For the desired effect,
set the fader just below the maximum.
</p>
`;

class Config extends UpdatingControlPanel<slpb.Config> {
    private _button: HTMLButtonElement;
    private _input: HTMLInputElement;

    constructor(cfg: Cfg) {
        super({ title: 'Configuration', help, data: cfg });

        this.innerHTML = `
<section class="grid grid-2-col">

<button>Capture Fader Threshold</button>
<input type="text" disabled />

</section>
`;

        this._button = this.querySelector('button');
        this._button.addEventListener('click', () => this._captureThreshold());
        this._input = this.querySelector('input');
    }

    update(cfg: slpb.Config) {
        console.log(`DEBUG CFG ${JSON.stringify(cfg)}`)
        this._input.value = cfg.faderThreshold.toString();
    }

    private _captureThreshold() {
        this._button.disabled = true;
        bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_REQUEST,
            type: slpb.MessageTypeRequest.CAPTURE_THRESHOLD_REQ,
            message: new slpb.CaptureThresholdRequest().toBinary(),
        })).then((reply) => {
            let threshold = slpb.CaptureThresholdResponse.fromBinary(reply.message).faderThreshold;
            let cfg = this.last.clone();
            cfg.faderThreshold = threshold;
            console.log(`DEBUG cap: ${threshold}`);
            this.save(cfg);
        }).finally(() => this._button.disabled = false);
    }
}
customElements.define('trackstar-stagelinq-config', Config, { extends: 'fieldset' });

export { Config };