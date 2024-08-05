class ColorPicker extends HTMLElement {
    onConfirm: (color: string) => void

    constructor() { 
        super();
        this.attachShadow({ mode: 'open' });
        this.shadowRoot!.innerHTML = `
<style>
</style>
<div id="colorSelect">
    <input type="color" id="colorPicker" />
    <button id="btn-confirm">&#x2705;</button>
    <button id="btn-cancel">No</button>
</div>
`;
    }

    connectedCallback() {
        this.shadowRoot!.querySelector("#btn-confirm")!
            .addEventListener('click',
                () => {
                if (this.onConfirm) {
                    let selectedColor = (this.shadowRoot!.querySelector('#colorPicker')! as HTMLInputElement).value;
                    this.onConfirm(selectedColor);
                }
                this.remove();
            });
        this.shadowRoot!.querySelector("#btn-cancel")!
            .addEventListener('click',
                () => {
                    this.remove();
                }
            );
    }
}

export { ColorPicker };