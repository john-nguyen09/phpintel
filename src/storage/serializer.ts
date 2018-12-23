import { TypeName } from "../type/name";
import { Location } from "../symbol/meta/location";
import { Range } from "../symbol/meta/range";
import { Position } from "../symbol/meta/position";
import { SymbolModifier } from "../symbol/meta/modifier";
import { TypeComposite } from "../type/composite";
import { App } from "../app";
import { LogWriter } from "../service/logWriter";

export class Serializer {
    public static readonly DEFAULT_SIZE = 1024;

    private buffer: Buffer;
    private offset: number;
    private length: number;

    constructor(buffer?: Buffer) {
        let logger = App.get<LogWriter>(LogWriter);

        if (buffer !== undefined) {
            if (typeof buffer == 'string') {
                logger.info(JSON.stringify(buffer));
            }

            this.buffer = buffer;
            this.length = buffer.length;
        } else {
            this.buffer = Buffer.alloc(Serializer.DEFAULT_SIZE);
            this.length = 0;
        }
        this.offset = 0;
    }

    private growHeap() {
        let newBuffer = Buffer.alloc(this.length + Serializer.DEFAULT_SIZE);
        this.buffer.copy(newBuffer);
        this.buffer = newBuffer;
    }

    private needs(noBytes: number) {
        if ((this.offset + noBytes) > this.buffer.length) {
            this.growHeap();
        }
    }

    public writeBuffer(buffer: Buffer) {
        this.needs(buffer.length);
        buffer.copy(this.buffer, this.offset);
        this.offset += buffer.length;
        this.length += buffer.length;
    }

    public readBuffer(): Buffer {
        return this.buffer.slice(this.offset);
    }

    public writeString(str: string) {
        let strBuffer = Buffer.from(str, 'utf8');

        this.writeInt32(strBuffer.length);
        this.writeBuffer(strBuffer);
    }

    public readString(): string {
        let length = this.readInt32();
        let strBuffer = this.buffer.slice(this.offset, this.offset + length);
        this.offset += length;

        return strBuffer.toString('utf8');
    }

    public writeInt32(value: number) {
        this.needs(4);
        this.buffer.writeInt32BE(value, this.offset);
        this.offset += 4;
        this.length += 4;
    }

    public readInt32(): number {
        let value = this.buffer.readInt32BE(this.offset);
        this.offset += 4;

        return value;
    }

    public writeBool(value: boolean) {
        this.needs(1);
        this.buffer.writeUInt8(value ? 1 : 0, this.offset);
        this.offset += 1;
        this.length += 1;
    }

    public readBool(): boolean {
        let value = this.buffer.readUInt8(this.offset) == 1 ? true : false;
        this.offset += 1;

        return value;
    }

    public writeTypeComposite(types: TypeComposite) {
        this.writeInt32(types.types.length);
        for (let type of types.types) {
            this.writeTypeName(type);
        }
    }

    public readTypeComposite(): TypeComposite {
        let numTypes = this.readInt32();
        let types = new TypeComposite();
        for (let i = 0; i < numTypes; i++) {
            types.push(this.readTypeName() || new TypeName(''));
        }

        return types;
    }

    public writeTypeName(name: TypeName | null) {
        if (name == null) {
            this.writeBool(false);
        } else {
            this.writeBool(true);
            this.writeString(name.getName());
        }
    }

    public readTypeName(): TypeName | null {
        let hasTypeName = this.readBool();

        if (!hasTypeName) {
            return null;
        }

        return new TypeName(this.readString());
    }

    public writeLocation(location: Location) {
        if (location.isEmpty) {
            this.writeBool(false);
        } else {
            this.writeBool(true);
            this.writeString(location.uri);
            this.writeRange(location.range);
        }
    }

    public readLocation(): Location {
        let hasLocation = this.readBool();
        
        if (!hasLocation) {
            return new Location();
        }

        return new Location(this.readString(), this.readRange());
    }

    public writeRange(range: Range) {
        this.writePosition(range.start);
        this.writePosition(range.end);
    }

    public readRange(): Range {
        return new Range(this.readPosition(), this.readPosition());
    }

    public writePosition(position: Position) {
        this.writeInt32(position.offset);
        this.writeInt32(position.line);
        this.writeInt32(position.character);
    }

    public readPosition(): Position {
        return new Position(this.readInt32(), this.readInt32(), this.readInt32());
    }

    public writeSymbolModifier(modifier: SymbolModifier) {
        this.writeInt32(modifier.getModifier());
        this.writeInt32(modifier.getVisibility());
    }

    public readSymbolModifier(): SymbolModifier {
        return new SymbolModifier(this.readInt32(), this.readInt32());
    }

    public getBuffer(): Buffer {
        return this.buffer.slice(0, this.length);
    }
}