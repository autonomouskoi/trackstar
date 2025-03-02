import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import { UpdatingControlPanel } from '/tk.js';
import { Cfg } from './controller.js';
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';
import { debounce } from "/debounce.js";

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

let help = document.createElement('div');
help.innerHTML = `
    <h3>Demo Delay Seconds</h3>
<p>
<code>trackstar</code> has a built-in demo mode to help while configuring it.
If <em>Demo Delay Seconds</em> is greater than 0, <code>trackstar</code> will
randomly pick from a built-in list of artist names and from a list of track
titles and send that as a newly-played track. It will wait the specified number
of seconds and repeat the process.
</p>

    <h3>Send Test Track</h3>
<p>
You may want to test how <code>trackstar</code> will handle some track data
but don't want to load up your DJ software and start playing. Using <em>Send</em>
Test Track</em> you can send a track through <code>trackstar</code> as if it had
come from DJ software.
</p>

<p>It will be processed exactly the same as a track that
did come from your DJ software, including any delay specified in <em>Track
Delay Seconds</em>. It will appear in your track log, in the overlay if you're
using <code>trackstaroverlay</code> and will appear in Twitch Chat if you're
using <code>trackstartwitchchat</code>.
</p>
`;

class Demo extends UpdatingControlPanel<tspb.Config> {
    private _demoDelay: HTMLInputElement;

    constructor(cfg: Cfg) {
        super({ title: 'Demo/Test', help, data: cfg });

        this.innerHTML = `
<section>
    <label for="input-demo-seconds"
        title="If greater than 0, send randomly-generated tracks this frequently"
    >
    Demo Delay Seconds</label>
<input type="number" id="input-demo-seconds" min="0" size="4"
        title="If greater than 0, send randomly-generated tracks this frequently"
    />
</section>

<details>
        <summary>Send Test Track</summary>
<form method="dialog">
    <div class="grid grid-2-col">

    <label for="input-test-artist">Artist</label>
    <input type="text" id="input-test-artist" />

    <label for="input-test-title">Title</label>
    <input type="text" id="input-test-title" />

    <input type="submit" value="Send"/>
    </div>
</form>
</details>
`;

        this._demoDelay = this.querySelector('#input-demo-seconds');
        this._demoDelay.addEventListener('change', debounce(1000, () => {
            let cfg = this.last.clone();
            cfg.demoDelaySeconds = parseInt(this._demoDelay.value);
            this.save(cfg);
        }));

        let testArtist: HTMLInputElement = this.querySelector('#input-test-artist');
        let testTitle: HTMLInputElement = this.querySelector('#input-test-title');
        this.querySelector('form').addEventListener('submit', () => {
            bus.send(new buspb.BusMessage({
                topic: TOPIC_REQUEST,
                type: tspb.MessageTypeRequest.SUBMIT_TRACK_REQ,
                message: new tspb.SubmitTrackRequest({
                    trackUpdate: new tspb.TrackUpdate({
                        deckId: 'Test',
                        track: new tspb.Track({
                            artist: testArtist.value,
                            title: testTitle.value,
                        }),
                        when: BigInt(new Date().getSeconds()),
                    }),
                }).toBinary(),
            }))
        });
    }

    update(cfg: tspb.Config) {
        this._demoDelay.value = cfg.demoDelaySeconds.toString();
    }
}
customElements.define('trackstar-demo', Demo, { extends: 'fieldset' });

export { Demo };