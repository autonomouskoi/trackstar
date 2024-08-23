import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tstc from "/m/trackstartwitchchat/pb/twitchchat_pb.js";

const TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_REQUEST);
const TOPIC_TRACKSTAR_TWITCHCHAT_COMMAND = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_COMMAND);

class Config extends HTMLElement {
    private _announcedButton: HTMLButtonElement;
    private _announceCheck: HTMLInputElement;
    private _config: tstc.Config;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.shadowRoot.innerHTML = `
<fieldset>
<legend>Twitch Chat Configuration</legend>

<label for="check-announce">Announce New Tracks</label>
<input id="check-announce" type="checkbox" />

<div>
    <button id="button-announce">Announce Current Track</button>
</div>
</fieldset>
`;
        this._announcedButton = this.shadowRoot.querySelector('#button-announce');
        this._announcedButton.onclick = () => this.announce();
        this._announceCheck = this.shadowRoot.querySelector('#check-announce');
        this._announceCheck.onchange = () => this.saveConfig();
        this.config = new tstc.Config();
        bus.waitForTopic(TOPIC_TRACKSTAR_TWITCHCHAT_REQUEST, 5000)
            .then(() => this.getConfig());
    }

    set config(config: tstc.Config) {
        this._config = config;
        this._announceCheck.checked = config.announce;
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
        let csr = new tstc.ConfigSetRequest();
        csr.config = this._config;
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_TRACKSTAR_TWITCHCHAT_COMMAND;
        msg.type = tstc.MessageTypeCommand.TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_REQ;
        msg.message = csr.toBinary();
        bus.sendWithReply(msg, (reply: buspb.BusMessage) => {
            if (reply.error) {
                throw reply.error;
            }
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