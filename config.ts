import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';
import { GloballyStyledHTMLElement } from '/global-styles.js';
import { debounce } from "/debounce.js";

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

class Config extends GloballyStyledHTMLElement {
    private _config = new tspb.Config();

    private _check_clear_bracketed_text: HTMLInputElement;
    private _dialog_new_replacement: HTMLDialogElement;
    private _div_replacements: HTMLDivElement;
    private _input_delay_seconds: HTMLInputElement;
    private _input_demo_seconds: HTMLInputElement;
    private _input_replacement_match: HTMLInputElement;
    private _input_replacement_artist: HTMLInputElement;
    private _input_replacement_title: HTMLInputElement;

    save: (config: tspb.Config) => void = () => { };

    constructor() {
        super();
        this.shadowRoot.innerHTML = `
<fieldset>
    <legend>Configuration</legend>

<div class="grid grid-2-col">
<label for="input-demo-seconds"
        title="If greater than 0, send randomly-generated tracks this frequently"
    >
    Demo Delay Seconds</label>
<input type="number" id="input-demo-seconds" min="0" size="4"
        title="If greater than 0, send randomly-generated tracks this frequently"
    />

<label for="input-delay-seconds"
        title="Delay sending tracks for this many seconds"
    >
    Track Delay Seconds</label>
<input type="number" id="input-delay-seconds" min="0" size="4"
        title="Delay sending tracks for this many seconds"
    />

<label for="check-clear-bracketed-text"
    title="Strip text from tracks inside square brackets, .e.g [delete this]"
    >Clear Bracketed Text</label>
<input type="checkbox" id="check-clear-bracketed-text"
    title="Strip text from tracks inside square brackets, .e.g [delete this]"
    />
</div>

<section>
    <h4>Track Replacements <button id="btn-new-replacement"> + </button></h4>
    <div class="grid grid-4-col" id="div-replacements"></div>
</section>

<dialog id="dialog-new-replacement">
<h2>New Track Replacement</h2>
<div class="grid grid-2-col">

<label for="input-replace-match">Match</label>
<input type="text" id="input-replace-match"
    required pattern="^.*[0-9a-zA-Z].*$"
/>

<label for="input-replace-artist">Artist</label>
<input type="text" id="input-replace-artist" />

<label for="input-replace-title">Title</label>
<input type="text" id="input-replace-title" />

<button id="btn-save-replacement">Save</button>
<button id="btn-cancel-replacement">Cancel</button>
</div>
</dialog>

<details>
        <summary>Send Test Track</summary>
<div class="grid grid-2-col">

<label for="input-test-artist">Artist</label>
<input type="text" id="input-test-artist" />

<label for="input-test-title">Title</label>
<input type="text" id="input-test-title" />

<button id="btn-test">Send</button>
</div>
</details>
</fieldset>
`;

        let onInput = debounce(1000, () => this._onSave());

        this._check_clear_bracketed_text = this.shadowRoot.querySelector('#check-clear-bracketed-text')
        this._check_clear_bracketed_text.addEventListener('input', () => onInput());
        this._div_replacements = this.shadowRoot.querySelector('#div-replacements');
        this._input_delay_seconds = this.shadowRoot.querySelector('#input-delay-seconds');
        this._input_delay_seconds.addEventListener('input', () => onInput());
        this._input_demo_seconds = this.shadowRoot.querySelector('#input-demo-seconds');
        this._input_demo_seconds.addEventListener('input', () => onInput());

        let buttonTest = this.shadowRoot.querySelector('#btn-test');
        buttonTest.addEventListener('click', () => this._sendTest());

        this._dialog_new_replacement = this.shadowRoot.querySelector('#dialog-new-replacement');
        let buttonReplacementSave = this.shadowRoot.querySelector('#btn-save-replacement') as HTMLButtonElement;
        let buttonNewReplacement = this.shadowRoot.querySelector('#btn-new-replacement') as HTMLButtonElement;
        this._input_replacement_match = this.shadowRoot.querySelector('#input-replace-match');
        this._input_replacement_match.addEventListener('input', () => {
            buttonReplacementSave.disabled = !this._input_replacement_match.validity.valid;
        });
        this._input_replacement_artist = this.shadowRoot.querySelector('#input-replace-artist');
        this._input_replacement_title = this.shadowRoot.querySelector('#input-replace-title');
        buttonNewReplacement.addEventListener('click', () => {
            this._input_replacement_match.value = '';
            this._input_replacement_artist.value = '';
            this._input_replacement_title.value = '';
            buttonReplacementSave.disabled = true;
            this._dialog_new_replacement.showModal()
        });
        let buttonReplacementCancel = this.shadowRoot.querySelector('#btn-cancel-replacement') as HTMLButtonElement;
        buttonReplacementCancel.addEventListener('click', () => this._dialog_new_replacement.close());
        buttonReplacementSave.addEventListener('click', () => {
            let track = new tspb.Track();
            track.artist = this._input_replacement_artist.value;
            track.title = this._input_replacement_title.value;
            let config = this._config.clone();
            config.trackReplacements[this._input_replacement_match.value] = track;
            this.save(config);
        });
    }

    set config(config: tspb.Config) {
        this._config = config;
        this._check_clear_bracketed_text.checked = config.clearBracketedText;
        this._input_delay_seconds.value = config.trackDelaySeconds.toString();
        this._input_demo_seconds.value = config.demoDelaySeconds.toString();

        if (Object.keys(this._config.trackReplacements).length) {
            this._div_replacements.innerHTML = `
<div class="column-header">Match</div>
<div class="column-header">Artist</div>
<div class="column-header">Title</div>
<div class="column-header"></div>
`;

            for (let match of Object.keys(this._config.trackReplacements)) {
                let matchDiv = document.createElement('div');
                matchDiv.innerHTML = `<code>${match}</code>`;
                this._div_replacements.appendChild(matchDiv);

                let track = this._config.trackReplacements[match];
                let artistDiv = document.createElement('div');
                artistDiv.innerText = track.artist;
                this._div_replacements.appendChild(artistDiv);

                let titleDiv = document.createElement('div');
                titleDiv.innerText = track.title;
                this._div_replacements.appendChild(titleDiv);

                let deleteButton = document.createElement('button') as HTMLButtonElement;
                deleteButton.innerText = 'Delete';
                deleteButton.addEventListener('click', () => {
                    let config = this._config.clone();
                    delete config.trackReplacements[match];
                    this.save(config);
                });
                this._div_replacements.appendChild(deleteButton);
            }

        }
    }

    private _onSave() {
        let config = this._config.clone();
        config.clearBracketedText = this._check_clear_bracketed_text.checked;
        config.demoDelaySeconds = parseInt(this._input_demo_seconds.value);
        config.trackDelaySeconds = parseInt(this._input_delay_seconds.value);

        this.save(config);
    }

    private _sendTest() {
        let artist = (this.shadowRoot.querySelector('#input-test-artist') as HTMLInputElement).value;
        let title = (this.shadowRoot.querySelector('#input-test-title') as HTMLInputElement).value;
        let str = new tspb.SubmitTrackRequest();
        str.trackUpdate = new tspb.TrackUpdate();
        str.trackUpdate.track = new tspb.Track();
        str.trackUpdate.deckId = 'Test';
        str.trackUpdate.track.artist = artist;
        str.trackUpdate.track.title = title;
        str.trackUpdate.when = BigInt(new Date().getSeconds());
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_REQUEST;
        msg.type = tspb.MessageTypeRequest.SUBMIT_TRACK_REQ;
        msg.message = str.toBinary();
        bus.sendWithReply(msg, () => { });
    }
}
customElements.define('trackstar-config', Config);

export { Config };