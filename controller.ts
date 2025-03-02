import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import { ValueUpdater } from "/vu.js";
import * as tspb from '/m/trackstar/pb/trackstar_pb.js';

const TOPIC_REQUEST = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_REQUEST);
const TOPIC_COMMAND = enumName(tspb.BusTopic, tspb.BusTopic.TRACKSTAR_COMMAND);

class Cfg extends ValueUpdater<tspb.Config> {
    constructor() {
        super(new tspb.Config());
    }

    refresh() {
        bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_REQUEST,
            type: tspb.MessageTypeRequest.CONFIG_GET_REQ,
            message: new tspb.ConfigGetRequest().toBinary(),
        })).then((reply) => {
            let cgResp = tspb.ConfigGetResponse.fromBinary(reply.message);
            this.update(cgResp.config);
        });
    }

    save(cfg: tspb.Config): Promise<void> {
        let csr = new tspb.ConfigSetRequest();
        csr.config = cfg;
        return bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_COMMAND,
            type: tspb.MessageTypeCommand.CONFIG_SET_REQ,
            message: csr.toBinary(),
        })).then((reply) => {
            let csResp = tspb.ConfigSetResponse.fromBinary(reply.message);
            this.update(csResp.config);
        });
    }
}
export { Cfg };