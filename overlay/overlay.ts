import * as deck from "./deck.js";
import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as overlaypb from "/m/trackstaroverlay/pb/overlay_pb.js"
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

const TOPIC_TRACKSTAR = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR);
const TOPIC_TRACKSTAR_OVERLAY_EVENT = enumName(overlaypb.BusTopic, overlaypb.BusTopic.TRACKSTAR_OVERLAY_EVENT);
const TOPIC_TRACKSTAR_OVERLAY_REQUEST =  enumName(overlaypb.BusTopic, overlaypb.BusTopic.TRACKSTAR_OVERLAY_REQUEST);

customElements.define('deck-track', deck.Deck);

let decksContainer = document.querySelector("#decksContainer");

let decksByID = {};
let tracksCount = 5;

fetch("/m/trackstar/build.json")
    .then(resp => resp.json())
    .then(jsonData => {
        interface BuildData {
            Software: string;
            Build: string;
        }
        let build = jsonData as BuildData;
        let deck = document.createElement("deck-track") as deck.Deck;
        decksContainer.appendChild(deck);
        deck.setTrack(build.Software, build.Build);
    });

const Finished = Symbol("finished");

class Animator {
    private _start: DOMHighResTimeStamp;
    private _durationMS: number;
    private _action: (inV: number | typeof Finished) => (typeof Finished | undefined);
    private _mapFn: (inV: number) => number;

    constructor(
        action: (inV: number | typeof Finished) => typeof Finished,
        durationMS: number,
        mapFn: (inV: number) => number = (n: number) => n,
    ) {
        this._action = action;
        this._durationMS = durationMS;
        this._mapFn = mapFn;

        window.requestAnimationFrame(x => this._animate(x));
    }

    private _step(current: DOMHighResTimeStamp): (number | typeof Finished) {
        if (this._start === undefined) {
            this._start = current;
        }
        let elapsed = current - this._start;
        if (elapsed > this._durationMS) {
            return Finished;
        }
        return elapsed / this._durationMS;
    }

    private _animate(current: DOMHighResTimeStamp) {
        let progress = this._step(current);
        if (this._action(progress) === Finished) {
            return;
        }
        if (progress === Finished) {
            return;
        }
        window.requestAnimationFrame(x => this._animate(x));
    }
}

let handleTrackstar = (msg: buspb.BusMessage) => {
    switch (msg.type) {
        case tspb.MessageType.TYPE_DECK_DISCOVERED:
        /*
        let dd = tspb.DeckDiscovered.fromBinary(msg.message);
        */
        case tspb.MessageType.TYPE_TRACK_UPDATE:
            let tu = tspb.TrackUpdate.fromBinary(msg.message);
            let deck = document.createElement("deck-track") as deck.Deck;
            deck.setTrack(tu.track!.artist, tu.track!.title);

            let newTrackAction = (progress: number | typeof Finished): typeof Finished => {
                if (progress === Finished) {
                    for (let i = 0;i < decksContainer.children.length;i++) {
                        let deck = decksContainer.children[i] as deck.Deck;
                        deck.setStyleProperty("transform", "translate(0, 0)");
                    }
                    return;
                }
                for (let i = 0;i < decksContainer.children.length;i++) {
                    let deck = decksContainer.children[i] as deck.Deck;
                    deck.setStyleProperty("transform", `translate(0, ${(1 - progress) * -100}%)`);
                }
                return;
            }
            new Animator(newTrackAction, 1000);
            if (decksContainer.children.length) {
                decksContainer.insertBefore(deck,decksContainer.children[0]);
            } else {
                decksContainer.appendChild(deck);
            }
            while (decksContainer.childNodes.length > tracksCount) {
                decksContainer.removeChild(decksContainer.lastChild);
            }
    }
}

let applyStyleUpdate = (su: overlaypb.StyleUpdate) => {
    let element = document.querySelector(su.selector) as HTMLElement;
    element.style.setProperty(su.property, su.value);
}

bus.subscribe(TOPIC_TRACKSTAR, handleTrackstar);
bus.subscribe(TOPIC_TRACKSTAR_OVERLAY_EVENT, (msg: buspb.BusMessage) => {
    switch (msg.type) {
        case overlaypb.MessageType.STYLE_UPDATE:
            let su = overlaypb.StyleUpdate.fromBinary(msg.message);
            applyStyleUpdate(su);
            break;
        case overlaypb.MessageType.TRACK_COUNT_UPDATE:
            let tcu = overlaypb.TrackCountUpdate.fromBinary(msg.message);
            tracksCount = tcu.count;
            break;
    }
})

let handleGetConfigReply = (msg: buspb.BusMessage) => {
    let gcr = overlaypb.GetConfigResponse.fromBinary(msg.message);
    gcr.config.styles.forEach(applyStyleUpdate);
    tracksCount = gcr.config.trackCount;
}

let getConfigMessage = new buspb.BusMessage();
getConfigMessage.topic = TOPIC_TRACKSTAR_OVERLAY_REQUEST;
getConfigMessage.type = overlaypb.MessageType.GET_CONFIG_REQUEST;
getConfigMessage.message = (new overlaypb.GetConfigRequest()).toBinary();
bus.waitForTopic(TOPIC_TRACKSTAR_OVERLAY_REQUEST,5000)
    .then(() => bus.sendWithReply(getConfigMessage, handleGetConfigReply));