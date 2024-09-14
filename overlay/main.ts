import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";

import * as overlaypb from "/m/trackstaroverlay/pb/overlay_pb.js";

const TOPIC_REQUEST = enumName(overlaypb.BusTopic, overlaypb.BusTopic.TRACKSTAR_OVERLAY_REQUEST);

function start(mainContainer: HTMLElement) {
    document.querySelector("title").innerText = 'Trackstar Overlay Customization';

    mainContainer.innerHTML = `
<h1>Trackstar Overlay Custom CSS</h1>
<textarea id="custom-css"
    style="display: block"
    cols="60"
    rows="20"
    autocorrect="off"
    autofocus="true"
    spellcheck="false"
></textarea>
<label for="color-picker">Handy Color Picker</label> <input type="color" id="color-picker"/>
<button>Save</button>
`;
    let customCSS = mainContainer.querySelector('textarea');
    // allow tabs in the textarea
    // https://stackoverflow.com/questions/6637341/use-tab-to-indent-in-textarea?page=1&tab=scoredesc#tab-top
    customCSS.addEventListener('keydown', (e) => {
        if (e.key != 'Tab') {
            return;
        }
        e.preventDefault();
        let start = customCSS.selectionStart;
        let end = customCSS.selectionEnd;
        customCSS.value = customCSS.value.substring(0, start) + '\t' + customCSS.value.substring(end);
        customCSS.selectionStart = customCSS.selectionEnd = start+1;
    });

    let colorLabel = mainContainer.querySelector('label') as HTMLLabelElement;
    let colorInput = mainContainer.querySelector('input') as HTMLInputElement;
    colorInput.addEventListener('change', () => {
        if (!navigator.clipboard) {
            return;
        } 
        navigator.clipboard.writeText(colorInput.value).then(() => {
            let originalText = colorLabel.innerText;
            colorLabel.innerText = 'Color copied!';
            setTimeout(() => {colorLabel.innerText = originalText}, 5000);
        });
    });

    let saveButton = mainContainer.querySelector('button');
    saveButton.disabled = true;

    saveButton.addEventListener('click', () => {
        saveButton.disabled = true;
        let csr = new overlaypb.ConfigSetRequest();
        csr.config = new overlaypb.Config();
        csr.config.customCss = customCSS.value;
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_REQUEST;
        msg.type = overlaypb.MessageType.CONFIG_SET_REQ;
        msg.message = csr.toBinary();
        bus.sendWithReply(msg, (reply) => {
            if (reply.error) {
                throw reply.error;
            }
            if (reply.type !== overlaypb.MessageType.CONFIG_SET_RESP) {
                return;
            }
            let csResp = overlaypb.ConfigSetResponse.fromBinary(msg.message);
            customCSS.value = csResp.config.customCss;
            saveButton.disabled = false;
        });
    });

    bus.waitForTopic(TOPIC_REQUEST, 5000)
        .then(() => {
            let msg = new buspb.BusMessage();
            msg.topic = TOPIC_REQUEST;
            msg.type = overlaypb.MessageType.GET_CONFIG_REQUEST;
            let gcr = new overlaypb.GetConfigRequest().toBinary();
            bus.sendWithReply(msg, (reply) => {
                if (reply.error) {
                    throw reply.error;
                }
                if (reply.type !== overlaypb.MessageType.GET_CONFIG_RESPONSE) {
                    return;
                }
                let gcResp = overlaypb.GetConfigResponse.fromBinary(reply.message);
                customCSS.value = gcResp.config.customCss;
                saveButton.disabled = false;
            });
        });
}

export { start };