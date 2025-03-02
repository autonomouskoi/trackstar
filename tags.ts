import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import { CfgUpdater, UpdatingControlPanel } from '/tk.js';
import { Cfg } from './controller.js';
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

let help = document.createElement('div');
help.innerHTML = `
<p>
The <code>Tags</code> feature allows you to define tags that can be applied to
the most recently received track. For example, a <em>banger</em> tag might indicate
a particularly good mix. A <em>fix</em> tag might indicate that you need to fix
the grid or hot cues of the track. This can keep you in the mix without forgetting
something important.
</p>

<p>
Below the <em>Tags</em> header is a table of the currently-defined tags. Clicking
the <code>Delete</code> button will delete the tag after a confirmation. Clicking
the <code>Apply</code> button will apply the tag to the most recent track. The
link to the right of the <code>Apply</code> button is a direct link to apply this
tag. This is useful with tools like a Stream Deck; a <em>Website</em> action can
use this URL with <code>GET request in background</code>.
</p>
`;

class Tags extends UpdatingControlPanel<tspb.Config> {
    private _tags: HTMLDivElement;

    constructor(cfg: Cfg) {
        super({ title: 'Tags', help, data: cfg });

        this.innerHTML = `
<div id="tags"></div>
<button>New Tag</button>
`;
        this._tags = this.querySelector('#tags');
        this._tags.classList.add('grid', 'grid-2-col');

        this.querySelector('button').addEventListener('click', () => this._newTag());
    }

    update(cfg: tspb.Config) {
        super.update(cfg);
        this.tags = cfg.tags;
    }

    set tags(tags: tspb.TrackTagConfig[]) {
        this._tags.textContent = '';
        tags.toSorted((a, b) => a.tag.localeCompare(b.tag))
            .forEach((tag) => addTagChip(this.updater, this._tags, tag));
    }

    private _newTag() {
        let name = prompt('New tag name?').trim();
        if (!name) {
            return;
        }
        if (this.last.tags.some((ttc) => ttc.tag == name)) {
            return;
        }
        let cfg = this.last.clone();
        cfg.tags.push(new tspb.TrackTagConfig({ tag: name }));
        this.update(cfg);
    }
}
customElements.define('trackstar-tags', Tags, { extends: 'fieldset' });

function addTagChip(cfg: CfgUpdater<tspb.Config>, parent: HTMLElement, ttc: tspb.TrackTagConfig) {
    let name = document.createElement('div');
    name.innerText = ttc.tag;

    let buttonsDiv = document.createElement('div');

    let del = document.createElement('button');
    del.innerText = 'Delete';
    del.addEventListener('click', () => {
        if (window.confirm(`Delete ${ttc.tag}?`)) {
            let newCfg = cfg.last.clone();
            newCfg.tags = newCfg.tags.filter((tag) => tag.tag != ttc.tag);
            cfg.save(newCfg);
        }
    })
    buttonsDiv.appendChild(del);

    let apply = document.createElement('button');
    apply.innerText = 'Apply';
    apply.addEventListener('click', () => bus.send(new buspb.BusMessage({
        topic: TOPIC_REQUEST,
        type: tspb.MessageTypeRequest.TAG_TRACK_REQ,
        message: new tspb.TagTrackRequest({
            tag: new tspb.TrackUpdateTag({
                tag: ttc.tag,
            })
        },
        ).toBinary(),
    })));
    buttonsDiv.appendChild(apply);

    let link: HTMLAnchorElement = document.createElement('a');
    link.innerHTML = '&#x1F517;';
    link.href = `/m/d6f95efeb3138d6e/_webhook?action=add_tag&tag=${ttc.tag}`;
    buttonsDiv.appendChild(link);

    parent.appendChild(name);
    parent.appendChild(buttonsDiv);
}

export { Tags };