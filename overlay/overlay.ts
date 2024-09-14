import { bus, enumName } from "/bus.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";
import * as overlaypb from "/m/trackstaroverlay/pb/overlay_pb.js";
import { TrackUpdate } from "./track.js";

const TOPIC_TRACKSTAR_EVENT = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_EVENT);
const TOPIC_OVERLAY_EVENT = enumName(overlaypb.BusTopic, tspb.BusTopic.TRACKSTAR_EVENT);

function start(mainContainer: HTMLDivElement) {
    document.querySelector("title").innerText = 'Trackstar Overlay';

    let tuElem = new TrackUpdate();
    mainContainer.appendChild(tuElem);
    bus.subscribe(TOPIC_TRACKSTAR_EVENT, (msg) => {
        if (msg.type !== tspb.MessageTypeEvent.TRACKSTAR_EVENT_TRACK_UPDATE) {
            return;
        }
        let tu = tspb.TrackUpdate.fromBinary(msg.message);
        tuElem.trackUpdate = tu;
    });

    let customCSSLink = document.querySelector('#custom-css-link') as HTMLLinkElement;
    bus.subscribe(TOPIC_OVERLAY_EVENT, (msg) => {
        if (msg.type === overlaypb.MessageType.CONFIG_UPDATED) {
            customCSSLink.href = `./custom-css?${Math.random()}`;
        }
    });
}



export { start };