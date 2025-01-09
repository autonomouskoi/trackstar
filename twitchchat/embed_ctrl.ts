import { bus, enumName } from "/bus.js";
import { GloballyStyledHTMLElement } from "/global-styles.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tstc from "/m/trackstartwitchchat/pb/twitchchat_pb.js";

const TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_REQUEST);
const TOPIC_TRACKSTAR_TWITCHCHAT_COMMAND = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_COMMAND);

class Config extends GloballyStyledHTMLElement {
    private _announceButton: HTMLButtonElement;
    private _announceCheck: HTMLInputElement;
    private _templateInput: HTMLInputElement;
    private _saveButton: HTMLButtonElement;
    private _config: tstc.Config;

    constructor() {
        super();
        this.shadowRoot.innerHTML = `
<fieldset>
<legend>Twitch Chat Configuration</legend>

<div class="grid grid-2-col">
<label for="check-announce">Announce New Tracks</label>
<input id="check-announce" type="checkbox" />

<label>Announce Template</label>
<div>
    <input id="input-announce-tmpl" type="text"
            size="50"
    />
    <button id="btn-save">Save</button>
</div>
</div>

<div>
    <button id="button-announce">Announce Current Track</button>
</div>
</fieldset>
`;
        this._announceButton = this.shadowRoot.querySelector('#button-announce');
        this._announceButton.addEventListener('click', () => this.announce());
        this._announceCheck = this.shadowRoot.querySelector('#check-announce');
        this._announceCheck.addEventListener('change', () => this.saveConfig());
        this._templateInput = this.shadowRoot.querySelector('#input-announce-tmpl');
        this._saveButton = this.shadowRoot.querySelector('#btn-save');
        this._saveButton.addEventListener('click', () => this.saveConfig());
        this.config = new tstc.Config();
        bus.waitForTopic(TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST, 5000)
            .then(() => this.getConfig());
    }

    set config(config: tstc.Config) {
        this._config = config;
        this._announceCheck.checked = config.announce;
        this._templateInput.value = config.template;
    }

    getConfig() {
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST;
        msg.type = tstc.MessageTypeRequest.TRACKSTAR_TWITCH_CHAT_CONFIG_GET_REQ;
        msg.message = new tstc.ConfigGetRequest().toBinary();
        bus.sendWithReply(msg, (reply: buspb.BusMessage) => {
            if (reply.error) {
                throw reply.error;
            }
            let cgr = tstc.ConfigGetResponse.fromBinary(reply.message);
            this.config = cgr.config;
        });
    }

    saveConfig() {
        this._config.announce = this._announceCheck.checked;
        this._config.template = this._templateInput.value;
        let csr = new tstc.ConfigSetRequest();
        csr.config = this._config;
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_TRACKSTAR_TWITCHCHAT_COMMAND;
        msg.type = tstc.MessageTypeCommand.TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_REQ;
        msg.message = csr.toBinary();
        bus.sendAnd(msg)
            .catch((err: buspb.Error) => {
                alert(err.detail);
            });
    }

    announce() {
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST;
        msg.type = tstc.MessageTypeRequest.TRACKSTAR_TWITCH_CHAT_TRACK_ANNOUNCE_REQ;
        msg.message = new tstc.TrackAnnounceRequest().toBinary();
        bus.sendWithReply(msg, (reply: buspb.BusMessage) => {
            if (reply.error) {
                throw reply.error;
            }
        });
    }
}
customElements.define('trackstar-twitchchat-config', Config);

function start(mainContainer: HTMLElement) {
    mainContainer.innerHTML = '<trackstar-twitchchat-config></trackstar-twitchchat-config>';
}

export { start };