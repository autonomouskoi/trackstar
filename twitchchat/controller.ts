import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as tstc from "/m/trackstartwitchchat/pb/twitchchat_pb.js";
import { ValueUpdater } from "/vu.js";

const TOPIC_REQUEST = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_REQUEST);
const TOPIC_COMMAND = enumName(tstc.BusTopics, tstc.BusTopics.TRACKSTAR_TWITCH_CHAT_COMMAND);

class Cfg extends ValueUpdater<tstc.Config> {
    constructor() {
        super(new tstc.Config());
    }

    refresh() {
        bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_REQUEST,
            type: tstc.MessageTypeRequest.TRACKSTAR_TWITCH_CHAT_CONFIG_GET_REQ,
            message: new tstc.ConfigGetRequest().toBinary(),
        })).then((reply) =>
            this.update(tstc.ConfigGetResponse.fromBinary(reply.message).config)
        );
    }

    save(cfg: tstc.Config): Promise<void> {
        return bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_COMMAND,
            type: tstc.MessageTypeCommand.TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_REQ,
            message: new tstc.ConfigSetRequest({ config: cfg }).toBinary(),
        })).then((reply) =>
            this.update(tstc.ConfigSetResponse.fromBinary(reply.message).config)
        );
    }
}

export { Cfg };