import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";

import * as config from "./config.js";
import * as log from "./log.js";

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);
const TOPIC_COMMAND = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_COMMAND);

function start(mainContainer: HTMLElement) {
    document.querySelector("title").innerText = 'Trackstar';
    mainContainer.innerText = 'Loading...';

    let cfg = new config.Config();
    let lg = new log.Log();

    let saveConfig = (config: tspb.Config) => {
        let csr = new tspb.ConfigSetRequest();
        csr.config = config;
        let msg = new buspb.BusMessage();
        msg.topic = TOPIC_COMMAND;
        msg.type = tspb.MessageTypeCommand.CONFIG_SET_REQ;
        msg.message = csr.toBinary();
        bus.sendWithReply(msg, (reply) => {
            if (reply.error) {
                throw reply.error;
            }
            if (reply.type !== tspb.MessageTypeCommand.CONFIG_SET_RESP) {
                return;
            }
            let csReply = tspb.ConfigSetResponse.fromBinary(reply.message);
            cfg.config = csReply.config;
        });
    };

    cfg.save = saveConfig;

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
                cfg.config = cgr.config;
                mainContainer.appendChild(cfg);
                mainContainer.appendChild(lg);
            });
        })
}

export { start };