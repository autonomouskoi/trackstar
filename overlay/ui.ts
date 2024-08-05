import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as overlaypb from "/m/trackstaroverlay/pb/overlay_pb.js";

const SELECT_MAIN_CONTAINER = '#mainContainer';

// https://stackoverflow.com/a/72207078
function debounce<T extends (...args: any[]) => void>(
    wait: number,
    callback: T,
    immediate = false,
) {
    // This is a number in the browser and an object in Node.js,
    // so we'll use the ReturnType utility to cover both cases.
    let timeout: ReturnType<typeof setTimeout> | null;

    return function <U>(this: U, ...args: Parameters<typeof callback>) {
        const context = this;
        const later = () => {
            timeout = null;

            if (!immediate) {
                callback.apply(context, args);
            }
        };
        const callNow = immediate && !timeout;

        if (typeof timeout === "number") {
            clearTimeout(timeout);
        }

        timeout = setTimeout(later, wait);

        if (callNow) {
            callback.apply(context, args);
        }
    };
}

function sendStyleUpdate(selector: string, property: string, value: string) {
    let su = new overlaypb.StyleUpdate();
    su.selector = selector;
    su.property = property;
    su.value = value;
    let msg = new buspb.BusMessage();
    msg.topic = enumName(overlaypb.BusTopic, overlaypb.BusTopic.TRACKSTAR_OVERLAY_REQUEST);
    msg.type = overlaypb.MessageType.SET_STYLE;
    msg.message = su.toBinary();
    bus.send(msg);

}

// ====== Box Scale
let sendScaleUpdate = debounce(100, (value: string) => {
    sendStyleUpdate(SELECT_MAIN_CONTAINER, "transform", `scale(${value})`);
});

let scaleSlider = document.querySelector("#scale-slider") as HTMLInputElement;
let scaleInput = document.querySelector("#scale-input") as HTMLInputElement;

let scaleSliderInput = (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    scaleInput.value = target.value;
    sendScaleUpdate(target.value);
}
scaleSlider.addEventListener("input", scaleSliderInput);
let scaleInputChange = (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    scaleSlider.value = target.value;
    sendScaleUpdate(target.value);
}
scaleInput.addEventListener("input", scaleInputChange);

scaleSlider.value = '1';
scaleInput.value = '1';


// ====== Text Scale
let sendTextScaleUpdate = debounce(100, (value: string) => {
    sendStyleUpdate(SELECT_MAIN_CONTAINER, "font-size", `${value}rem`);
});

let textScaleSlider = document.querySelector("#text-scale-slider") as HTMLInputElement;
let textScaleInput = document.querySelector("#text-scale-input") as HTMLInputElement;

let textScaleSliderInput = (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    textScaleInput.value = target.value;
    sendTextScaleUpdate(target.value);
}
textScaleSlider.addEventListener("input", textScaleSliderInput);
let textScaleInputChange = (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    textScaleSlider.value = target.value;
    sendTextScaleUpdate(target.value);
}
textScaleInput.addEventListener("input", textScaleInputChange);

textScaleSlider.value = '1';
textScaleInput.value = '1';

// ======= Text Color
let textColorPicker = document.querySelector("#text-color") as HTMLInputElement;
let sendTextColorUpdate = debounce(100, (value: string) => {
    sendStyleUpdate(SELECT_MAIN_CONTAINER, "color", value);
})
textColorPicker.addEventListener("input", (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    sendTextColorUpdate(target.value);
})

// ======= Text Color
let textOutlineColorPicker = document.querySelector("#text-outline-color") as HTMLInputElement;
let sendTextOutlineColorUpdate = debounce(100, (value: string) => {
    sendStyleUpdate(SELECT_MAIN_CONTAINER, "-webkit-text-stroke-color", value) ;
});
textOutlineColorPicker.addEventListener("input", (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    sendTextOutlineColorUpdate(target.value);
});

// ======= Text Width
let textOutlineWidthInput = document.querySelector("#text-outline-width") as HTMLInputElement;
let sendTextOutlineWidthUpdate = debounce(100, (value: string) => {
    sendStyleUpdate(SELECT_MAIN_CONTAINER, "-webkit-text-stroke-width", `${value}px`);
});
textOutlineWidthInput.addEventListener("input", (ev: Event) => {
    let target = ev.target as HTMLInputElement;
    sendTextOutlineWidthUpdate(target.value);
});


let handleGetConfigReply = (msg: buspb.BusMessage) => {
    let gcr = overlaypb.GetConfigResponse.fromBinary(msg.message);
    gcr.config.styles.forEach((su: overlaypb.StyleUpdate) => {
        switch (su.selector) {
            case SELECT_MAIN_CONTAINER:
                switch (su.property) {
                    case 'font-size':
                       let fontSizeMatch = su.value.match(/([0-9.]+)rem/);
                       if (fontSizeMatch) {
                            textScaleSlider.value = fontSizeMatch[1];
                            textScaleInput.value = fontSizeMatch[1];
                       }
                       break
                    case 'transform':
                       let transformMatch = su.value.match(/scale\(([0-9.]+)\)/);
                       if (transformMatch) {
                            scaleSlider.value = transformMatch[1];
                            scaleInput.value = transformMatch[1];
                       }
                       break
                    case 'color':
                       textColorPicker.value = su.value; 
                       break
                    case '-webkit-text-stroke-color':
                        textOutlineColorPicker.value = su.value;
                       break
                    case '-webkit-text-stroke-width':
                        let textOutlineWidthMatch = su.value.match(/([0-9])px/);
                        if (textOutlineWidthMatch) {
                            textOutlineWidthInput.value = textOutlineWidthMatch[1];
                        }
                       break
                }
                break;
        }
    });
}
let getConfigMessage = new buspb.BusMessage();
getConfigMessage.topic = enumName(overlaypb.BusTopic, overlaypb.BusTopic.TRACKSTAR_OVERLAY_REQUEST);
getConfigMessage.type = overlaypb.MessageType.GET_CONFIG_REQUEST;
getConfigMessage.message = (new overlaypb.GetConfigRequest()).toBinary();
setTimeout(() => bus.sendWithReply(getConfigMessage, handleGetConfigReply), 250);