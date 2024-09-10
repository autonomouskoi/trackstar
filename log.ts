import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

import { GloballyStyledHTMLElement } from '/global-styles.js';

const TOPIC_EVENT = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_EVENT);
const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

// let mainContainer = document.querySelector('#mainContainer')

interface entry {
    when: Date;
    artist: string;
    title: string;
}

//let entries: entry[] = new Array();

/*
let handleTrackstar = (msg: buspb.BusMessage) => {
    if (msg.type !== tspb.MessageTypeEvent.TRACKSTAR_EVENT_TRACK_UPDATE) {
        return;
    }
    let tu = tspb.TrackUpdate.fromBinary(msg.message);
    let newEntry: entry = {
        when: new Date(Number(tu.when) * 1000),
        artist: tu.track.artist,
        title: tu.track.title,
    }
    entries.push(newEntry);
    let when = document.createElement("div");
    when.innerText = newEntry.when.toLocaleTimeString();
    let artist = document.createElement("div");
    artist.innerText = newEntry.artist;
    let title = document.createElement("div");
    title.innerText = newEntry.title;

    if (mainContainer.children.length) {
        mainContainer.insertBefore(title, mainContainer.children[0]);
        mainContainer.insertBefore(artist, mainContainer.children[0]);
        mainContainer.insertBefore(when, mainContainer.children[0]);
    } else {
        mainContainer.appendChild(when);
        mainContainer.appendChild(artist);
        mainContainer.appendChild(title);
    }
}
    */

/*
let download = () => {
    let data = entries.map(entry => {
        return `${JSON.stringify(entry.when)},${JSON.stringify(entry.artist)},${JSON.stringify(entry.title)}`
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

(document.querySelector('#download-button') as HTMLButtonElement).onclick = download;
*/

class Log extends GloballyStyledHTMLElement {
    private _table: HTMLTableElement;

    private _entries: entry[] = [];

    constructor() {
        super();
        this.shadowRoot.innerHTML = `
<style>
td {
    padding: 0.25rem;
}
</style>
<fieldset>
<legend>Track Log <button id="btn-download">Download</button></legend>
<table>
    <tr>
        <th>Time</th>
        <th>Artist</th>
        <th>Title</th>
    </tr>
</table>
</fieldset>
`;
        this._table = this.shadowRoot.querySelector('table')
        let dlButton = this.shadowRoot.querySelector('#btn-download');
        dlButton.addEventListener('click', () => this._onDownload());
    }

    connectedCallback() {
        bus.subscribe(TOPIC_EVENT, (msg: buspb.BusMessage) => this._handleEvent(msg));
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_REQUEST;
        msg.type = tspb.MessageTypeRequest.GET_ALL_TRACKS_REQ;
        msg.message = new tspb.GetAllTracksRequest().toBinary();
        bus.sendWithReply(msg, (reply) => {
            if (reply.error) {
                throw reply.error;
            }
            if (reply.type !== tspb.MessageTypeRequest.GET_ALL_TRACKS_RESP) {
                return;
            }
            let gatResp = tspb.GetAllTracksResponse.fromBinary(reply.message);
            gatResp.tracks.forEach((tu) => this._addTrack(tu));
        });
    }

    private _addTrack(tu: tspb.TrackUpdate) {
        let newEntry: entry = {
            when: new Date(Number(tu.when) * 1000),
            artist: tu.track.artist,
            title: tu.track.title,
        }
        this._entries.push(newEntry);
        let when = document.createElement("td");
        when.innerText = newEntry.when.toLocaleTimeString();
        let artist = document.createElement("td");
        artist.innerText = newEntry.artist;
        let title = document.createElement("td");
        title.innerText = newEntry.title;

        let row = document.createElement('tr') as HTMLTableRowElement;
        row.appendChild(when);
        row.appendChild(artist);
        row.appendChild(title);
        this._table.appendChild(row);
    }

    private _handleEvent(msg: buspb.BusMessage) {
        if (msg.type !== tspb.MessageTypeEvent.TRACKSTAR_EVENT_TRACK_UPDATE) {
            return;
        }
        let tu = tspb.TrackUpdate.fromBinary(msg.message);
        this._addTrack(tu);
    }

    private _onDownload() {
        let data = this._entries.map(entry => {
            return `${JSON.stringify(entry.when)},${JSON.stringify(entry.artist)},${JSON.stringify(entry.title)}`
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