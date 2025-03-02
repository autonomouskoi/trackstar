import { UpdatingControlPanel } from '/tk.js';
import { Cfg } from './controller.js';
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';
import { debounce } from '/debounce.js';

let help = document.createElement('div');
help.innerHTML = `
    <h3>Track Delay Seconds</h3>
<p>
When <code>trackstar</code> is notified that a new track is playing it will
delay notifying other modules by this number of seconds. When set to 0, there's
no delay. A value of 5 means that nothing else will become aware of the new
track for at least 5 seconds.
</p>

    <h3>Save Sessions</h3>

<p>
When <code>Save Sessions</code> is checked, <code>trackstar</code> will save
a record of the tracks played in this session to the database to be recalled
later.
</p>
`;

class General extends UpdatingControlPanel<tspb.Config> {
    private _delay: HTMLInputElement;
    private _save: HTMLInputElement;

    constructor(cfg: Cfg) {
        super({ title: 'General', help, data: cfg });

        this.innerHTML = `
<section class="grid grid-2-col">
<label for="input-delay-seconds"
        title="Delay sending tracks for this many seconds"
    >
    Track Delay Seconds</label>
<input type="number" id="input-delay-seconds" min="0" size="4"
        title="Delay sending tracks for this many seconds"
    />

<label for="check-save-sessions"
        title="Save sessions to the database"
    >
    Save Sessions</label>
<input type="checkbox" id="check-save-sessions"
        title="Save sessions to the database"
    />
</section>
`;

        this._delay = this.querySelector('#input-delay-seconds');
        this._delay.addEventListener('change', debounce(1000, () => {
            let cfg = this.last.clone();
            cfg.trackDelaySeconds = parseInt(this._delay.value);
            this.save(cfg);
        }))

        this._save = this.querySelector('#check-save-sessions');
        this._save.addEventListener('change', () => {
            let cfg = this.last.clone();
            cfg.saveSessions = this._save.checked;
            this.save(cfg);
        });
    }

    update(cfg: tspb.Config) {
        this._delay.value = cfg.trackDelaySeconds.toString();
        this._save.checked = cfg.saveSessions;
    }
}
customElements.define('trackstar-general', General, { extends: 'fieldset' });

export { General };