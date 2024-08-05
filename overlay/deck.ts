class Deck extends HTMLElement {
    private _artist = 'Artist';
    private _title  = 'Title';

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.shadowRoot!.innerHTML = `
<div id="deck">
        <div id="artist">${this._artist}</div>
        <div id="title">${this._title}</div>
</div>
`;
/*
        this.shadowRoot!.innerHTML = `
<style>
#deck {
    border: solid black 1px;
    width: 208px;
}
</style>
<div id="deck">
        <div id="artist">${this._artist}</div>
        <div id="title">${this._title}</div>
</div>
`;
*/
    }

    setStyleProperty(property: string, value: string) {
        (this.shadowRoot!.querySelector("#deck") as HTMLElement).style.setProperty(property, value);
    }

    setTrack(artist: string, title: string) {
        this._artist = artist;
        this._title = title;
        (this.shadowRoot!.querySelector("#artist") as HTMLElement).innerText = this._artist;
        (this.shadowRoot!.querySelector("#title") as HTMLElement).innerText = this._title;
    }
}

export { Deck };