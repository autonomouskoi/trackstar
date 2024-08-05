import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

let mainContainer = document.querySelector('#mainContainer')

interface entry {
    when: Date;
    artist: string;
    title: string;
}

let entries: entry[] = new Array();

let handleTrackstar = (msg: buspb.BusMessage) => {
    if (msg.type !== tspb.MessageType.TYPE_TRACK_UPDATE) {
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

bus.subscribe(enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR), handleTrackstar);