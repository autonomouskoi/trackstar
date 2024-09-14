import * as tspb from '/m/trackstar/pb/trackstar_pb.js';

class TrackUpdate extends HTMLElement {
    private _dev_track: HTMLDivElement;

    constructor() {
        super();
        this.innerHTML = `
<div id="track" class="track"></div>
`;
        this._dev_track = this.querySelector('#track');
    }

    set trackUpdate (tu: tspb.TrackUpdate) {
        this._dev_track.classList.remove('fadeIn');
        this._dev_track.classList.add('fadeOut');
        this._dev_track.addEventListener('animationend', () => {
            let when = new Date(Number(tu.when) * 1000);
            this._dev_track.innerHTML = `
<div class="deck-id">${tu.deckId}</div>
<time class="when">${when}</time>
<div class="artist">${tu.track.artist}</div>
<div class="title">${tu.track.title}</div>
`;
            this._dev_track.classList.remove('fadeOut');
            this._dev_track.classList.add('fadeIn');
        }, { once: true });
    }
}
customElements.define('trackstar-overlay-track-update', TrackUpdate);

export { TrackUpdate };