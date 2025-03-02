import { UpdatingControlPanel } from '/tk.js';
import { Cfg } from './controller.js';
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';

let help = document.createElement('div');
help.innerHTML = `
    <h3>Clear Bracketed Text</h3>
<p>
You may want to hide specific details of a track you play. For example, it
might be a yet unreleased track and while you have permission to play it, you
may not have permission to tell people about it. The <code>Clear Bracketed Text</code>
setting will cause <code>trackstar</code> to process the Artist and Title of
each track. Any text contained in square brackets (<code>[]</code>) will be
replaced with a space. Then, duplicate spaces will be removed.
</p>

<p>
Say you've received the unreleased track by <em>Billain't</em> entitled
<em>Smacula</em>. You can edit the track data in your DJ software to have the
Artist read <code>Neurofunk Producer [Billain't]</code> and the title to read
<code>Unreleased Track [Smacula]</code>. When the track is played with
<em>Clear Bracketed Text</em> enabled <code>trackstar</code> will edit the
track message to have the artist <code>Neurofunk Producer</code> and the title
<code>Unreleased Track</code>. When the track is released you can go back into
your DJ software and strip out everything outside the brackets and the brackets
themselves to have the original text displayed with future plays.
</p>

    <h3>Track Replacements</h3>
<p>
The <em>Track Replacements</em> feature serves a similar purpose to <em>Clear
Bracketed Text</em> but works in a different way. Each <em>replacement</em> has
three parts: <em>Match</em>, <em>Artist</em>, and <em>Title</em>. When a track
is played where the Artist <em>or</em> Title matches the text of <em>Match</em>,
<code>trackstar</code> will replace the Artist and Title of the track message
with the <em>Artist</em> and <em>Title</em> specified in the Track Replacement.
</p>

<p>
For example, you can create a track replacement like the following:
</p>
<dl>
    <dt>Match</dt>
        <dd>-UR-</dd>
    <dt>Artist</dt>
        <dd>Artist Redacted</dd>
    <dt>Title</dt>
        <dd>Unreleased Track</dd>
</dl>
<p>
You can then update your library so that <code>-UR-</code> appears in the Artist
or Title of all unreleased tracks. When you play one of these tracks, the track
data will be replaced to show <code>Artist Redacted</code>/<code>Unreleased Track</code>.
This requires less editing in your library but is less flexible.
</p>

<p>
Note: The <em>Match</em> is <em>case-sensitive</em>. <code>-UR-</code> as your
match will effect tracks with <code>-UR-</code> in them, but not <code>-Ur-</code>
or <code>-ur-</code>.
</p>
`;

class Replace extends UpdatingControlPanel<tspb.Config> {
    private _clearCheck: HTMLInputElement;
    private _replacements: HTMLDivElement;
    private _newDialog: HTMLDialogElement;

    constructor(cfg: Cfg) {
        super({ title: 'Track Replacement', help, data: cfg });

        this.innerHTML = `
<section>
    <label for="clear-check"
        title="Strip text from tracks inside square brackets, .e.g [delete this]"
    >Clear Bracketed Text</label>
    <input id="clear-check" type="checkbox"
        title="Strip text from tracks inside square brackets, .e.g [delete this]"
    />
</section>

<section>
    <h4>Track Replacements <button id="btn-new-replacement"> + </button></h4>
    <div class="grid grid-4-col" id="div-replacements"></div>
</section>

<dialog>
    <h2>New Track Replacement</h2>
    <div class="grid grid-2-col">

    <label for="input-replace-match">Match</label>
    <input type="text" id="input-replace-match"
        required pattern="^.*[0-9a-zA-Z].*$"
    />

    <label for="input-replace-artist">Artist</label>
    <input type="text" id="input-replace-artist" />

    <label for="input-replace-title">Title</label>
    <input type="text" id="input-replace-title" />

    <button id="btn-save-replacement">Save</button>
    <button id="btn-cancel-replacement">Cancel</button>
    </div>
</dialog>
`;

        this._clearCheck = this.querySelector('#clear-check');
        this._clearCheck.addEventListener('click', () => this._updateClearBracketedText());
        this._replacements = this.querySelector('#div-replacements');
        this._newDialog = this.querySelector('dialog');

        this._wireNewReplacementDialog();
    }

    update(cfg: tspb.Config) {
        super.update(cfg);
        this._clearCheck.checked = cfg.clearBracketedText;
        this._displayReplacements(cfg.trackReplacements);
        this._newDialog.close();
    }

    private _wireNewReplacementDialog() {
        let saveBtn: HTMLButtonElement = this._newDialog.querySelector('#btn-save-replacement');

        let match: HTMLInputElement = this._newDialog.querySelector('#input-replace-match');
        match.addEventListener('input', () => saveBtn.disabled = !match.validity.valid);
        let artist: HTMLInputElement = this._newDialog.querySelector('#input-replace-artist');
        let title: HTMLInputElement = this._newDialog.querySelector('#input-replace-title');

        let newBtn: HTMLButtonElement = this.querySelector('#btn-new-replacement');
        newBtn.addEventListener('click', () => {
            [match, artist, title].forEach((v) => v.value = '');
            saveBtn.disabled = true;
            this._newDialog.showModal();
        });

        let cancelBtn: HTMLButtonElement = this._newDialog.querySelector('#btn-cancel-replacement');
        cancelBtn.addEventListener('click', () => this._newDialog.close());

        saveBtn.addEventListener('click', () => {
            let track = new tspb.Track();
            track.artist = artist.value;
            track.title = title.value;
            let config = this.last.clone();
            config.trackReplacements[match.value] = track;
            this.save(config);
        });
    }

    private _displayReplacements(replacements: { [key: string]: tspb.Track }) {
        this._replacements.innerHTML = `
<div class="column-header">Match</div>
<div class="column-header">Artist</div>
<div class="column-header">Title</div>
<div class="column-header"></div>
`;

        for (let match of Object.keys(replacements).toSorted()) {
            let matchDiv = document.createElement('div');
            matchDiv.innerHTML = `<code>${match}</code>`;
            this._replacements.appendChild(matchDiv);

            let track = replacements[match];
            let artistDiv = document.createElement('div');
            artistDiv.innerText = track.artist;
            this._replacements.appendChild(artistDiv);

            let titleDiv = document.createElement('div');
            titleDiv.innerText = track.title;
            this._replacements.appendChild(titleDiv);

            let deleteButton = document.createElement('button') as HTMLButtonElement;
            deleteButton.innerText = 'Delete';
            deleteButton.addEventListener('click', () => {
                if (!confirm(`Delete ${match}?`)) {
                    return;
                }
                let config = this.last.clone();
                delete config.trackReplacements[match];
                this.save(config);
            });
            this._replacements.appendChild(deleteButton);
        }
    }

    private _updateClearBracketedText() {
        let cfg = this.last.clone();
        cfg.clearBracketedText = this._clearCheck.checked;
        this.save(cfg);
    }
}
customElements.define('trackstar-replace', Replace, { extends: 'fieldset' });

export { Replace };