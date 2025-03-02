import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

import * as log from "./log.js";
import { Cfg } from "./controller.js";
import { General } from "./general.js";
import { Demo } from "./demo.js";
import { Replace } from "./replace.js";
import { Tags } from "./tags.js";

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);

function start(mainContainer: HTMLElement) {
    let lg = new log.Log();
    let cfg = new Cfg();

    bus.waitForTopic(TOPIC_REQUEST, 5000)
        .then(() => {
            let msg = new buspb.BusMessage();
            msg.topic = TOPIC_REQUEST;
            msg.type = tspb.MessageTypeRequest.CONFIG_GET_REQ;
            msg.message = new tspb.ConfigGetRequest().toBinary();
            bus.sendWithReply(msg, (reply: buspb.BusMessage) => {
                if (reply.error) {
                    throw reply.error;
                }
                mainContainer.textContent = '';
                let cgr = tspb.ConfigGetResponse.fromBinary(reply.message);
                mainContainer.appendChild(new General(cfg));
                mainContainer.appendChild(new Demo(cfg));
                mainContainer.appendChild(new Replace(cfg));
                mainContainer.appendChild(new Tags(cfg));
                mainContainer.appendChild(lg);
                cfg.refresh();
            });
        })
}

export { start };