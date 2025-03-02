import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

import { GloballyStyledHTMLElement } from '/global-styles.js';

const TOPIC_EVENT = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_EVENT);
const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

// an entry holds a track update for display and rendering as CSV
interface entry {
    when: Date;
    artist: string;
    title: string;
    deck: string;
    tags: string;
}

class Log extends GloballyStyledHTMLElement {
    private _table: HTMLTableElement;
    private _session_select: HTMLSelectElement;

    private _current_session = '';
    private _entries: entry[] = [];

    constructor() {
        super();
        this.shadowRoot.innerHTML = `
<style>
legend {
    font-weight: bold;
    font-size: 1.3rem;
}
td {
    padding: 0.25rem;
}
</style>
<fieldset>
<legend>Track Log
    <select id="select-session"></select>
    <button id="btn-download">Download</button>
</legend>
<table></table>
</fieldset>
`;
        this._table = this.shadowRoot.querySelector('table')

        this._session_select = this.shadowRoot.querySelector('#select-session');
        this._session_select.addEventListener('change', () => {
            // when the selection changes, grab chosen session and display it
            bus.sendWithReply(new buspb.BusMessage({
                topic: TOPIC_REQUEST,
                type: tspb.MessageTypeRequest.GET_SESSION_REQ,
                message: new tspb.GetSessionRequest({
                    session: BigInt(this._session_select.value),
                }).toBinary(),
            }), (reply) => {
                let resp = tspb.GetSessionResponse.fromBinary(reply.message);
                this.session = resp.session;
            });
        });

        let dlButton = this.shadowRoot.querySelector('#btn-download');
        dlButton.addEventListener('click', () => this._onDownload());
    }

    connectedCallback() {
        bus.subscribe(TOPIC_EVENT, (msg: buspb.BusMessage) => this._handleEvent(msg));
        bus.sendWithReply(new buspb.BusMessage({ // request the available sessions
            topic: TOPIC_REQUEST,
            type: tspb.MessageTypeRequest.LIST_SESSIONS_REQ,
            message: new tspb.ListSessionsRequest().toBinary(),
        }), (reply) => {
            let resp = tspb.ListSessionsResponse.fromBinary(reply.message);
            this._session_select.textContent = '';
            // create a select option for each session in descending order
            resp.sessions.reverse().forEach((session) => {
                this._current_session = this._current_session ? this._current_session : session.toString();
                let option: HTMLOptionElement = document.createElement('option');
                option.value = session.toString();
                option.innerText = new Date(Number(session) * 1000).toLocaleString();
                this._session_select.appendChild(option);
            });
        });
        this._getSession(BigInt(0));
    }

    // retrieve and display a specific session
    private _getSession(sessionID: bigint) {
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_REQUEST;
        msg.type = tspb.MessageTypeRequest.GET_SESSION_REQ;
        msg.message = new tspb.GetSessionRequest(
            { session: sessionID },
        ).toBinary();
        bus.sendWithReply(msg, (reply) => {
            if (reply.error) {
                throw reply.error;
            }
            if (reply.type !== tspb.MessageTypeRequest.GET_SESSION_RESP) {
                return;
            }
            let gatResp = tspb.GetSessionResponse.fromBinary(reply.message);
            this.session = gatResp.session;
        });
    }

    // setting the session renders it
    private set session(session: tspb.Session) {
        this._initTable();
        this._entries = [];
        session.tracks.forEach((tu) => this._addTrack(tu));
    }

    // reset the table to display a whole new session
    private _initTable() {
        this._table.innerHTML = `
    <tr>
        <th>Time</th>
        <th>Artist</th>
        <th>Title</th>
        <th>Deck</th>
        <th>Tags</th>
    </tr>`;
    }

    // add a track to the currently displayed session
    private _addTrack(tu: tspb.TrackUpdate) {
        let newEntry: entry = {
            when: new Date(Number(tu.when) * 1000),
            artist: tu.track.artist,
            title: tu.track.title,
            deck: tu.deckId,
            // TODO: figure out a better way to display the tags
            tags: tu.tags.map((tag) => tag.tag).join(','),
        }
        this._entries.push(newEntry);
        let when = document.createElement('td');
        when.innerText = newEntry.when.toLocaleTimeString();
        let artist = document.createElement('td');
        artist.innerText = newEntry.artist;
        let title = document.createElement('td');
        title.innerText = newEntry.title;
        let deck = document.createElement('td');
        deck.innerText = newEntry.deck;
        let tags = document.createElement('td');
        tags.innerText = newEntry.tags;

        let row = document.createElement('tr') as HTMLTableRowElement;
        row.appendChild(when);
        row.appendChild(artist);
        row.appendChild(title);
        row.appendChild(deck);
        row.appendChild(tags);
        this._table.appendChild(row);
    }

    private _handleEvent(msg: buspb.BusMessage) {
        switch (msg.type) {
            case tspb.MessageTypeEvent.TRACKSTAR_EVENT_TRACK_UPDATE:
                if (this._current_session != this._session_select.selectedOptions[0].value) {
                    // only append a track update if we're displaying the current session
                    return;
                }
                let tu = tspb.TrackUpdate.fromBinary(msg.message);
                this._addTrack(tu);
                break;
            case tspb.MessageTypeEvent.TRACKSTAR_EVENT_SESSION_UPDATE:
                // the entire session should be redisplayed
                this._getSession(BigInt(0));
                break;
        }
    }

    private _onDownload() {
        // JSON.stringify quotes and escapes quotes for strings
        let data = this._entries.map(entry => {
            return `${JSON.stringify(entry.when)},${JSON.stringify(entry.artist)},${JSON.stringify(entry.title)},${JSON.stringify(entry.deck)},${JSON.stringify(entry.tags)}`
        }).join('\n');
        let file = new File([data], 'trackstar.csv', {
            type: 'text/csv',
        });
        let link = document.createElement('a');
        let url = URL.createObjectURL(file);
        link.href = url;
        link.download = file.name;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
    }
}
customElements.define('trackstar-log', Log);

export { Log };