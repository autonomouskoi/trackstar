import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as slpb from '/m/trackstarstagelinq/pb/stagelinq_pb.js';
import { ValueUpdater } from "/vu.js";

const TOPIC_REQUEST = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_REQUEST);
const TOPIC_COMMAND = enumName(slpb.BusTopics, slpb.BusTopics.STAGELINQ_COMMAND);

class Cfg extends ValueUpdater<slpb.Config> {
    constructor() {
        super(new slpb.Config());
    }

    refresh() {
        bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_REQUEST,
            type: slpb.MessageTypeRequest.CONFIG_GET_REQ,
            message: new slpb.ConfigGetRequest().toBinary(),
        })).then((reply) => {
            let cgResp = slpb.ConfigGetResponse.fromBinary(reply.message);
            this.update(cgResp.config);
        });
    }

    save(cfg: slpb.Config): Promise<void> {
        let csr = new slpb.ConfigSetRequest();
        csr.config = cfg;
        return bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_COMMAND,
            type: slpb.MessageTypeCommand.CONFIG_SET_REQ,
            message: csr.toBinary(),
        })).then((reply) => {
            let csResp = slpb.ConfigSetResponse.fromBinary(reply.message);
            this.update(csResp.config);
        });
    }
}
export { Cfg };