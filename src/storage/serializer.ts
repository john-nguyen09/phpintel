import { TypeName } from "../type/name";
import { Location } from "../symbol/meta/location";
import { Range } from "../symbol/meta/range";
import { SymbolModifier } from "../symbol/meta/modifier";
import { TypeComposite } from "../type/composite";
import { NamespaceName } from "../symbol/name/namespaceName";

export class Serializer {
    public static readonly DEFAULT_SIZE = 1024;

    private buffer: Buffer;
    private offset: number;
    private length: number;

    constructor(buffer?: Buffer) {
        if (buffer !== undefined) {
            this.buffer = buffer;
            this.length = buffer.length;
        } else {
            this.buffer = Buffer.alloc(Serializer.DEFAULT_SIZE);
            this.length = 0;
        }
        this.offset = 0;
    }

    private growHeap(proposedSize: number) {
        let newSize = Math.max(this.length + Serializer.DEFAULT_SIZE, proposedSize);
        let newBuffer = Buffer.alloc(newSize);
        this.buffer.copy(newBuffer);
        this.buffer = newBuffer;
    }

    private needs(noBytes: number) {
        if ((this.offset + noBytes) > this.buffer.length) {
            this.growHeap(this.offset + noBytes);
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
            this.writeString(name.name);
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
        this.writeInt32(range.start);
        this.writeInt32(range.end);
    }

    public readRange(): Range {
        return new Range(this.readInt32(), this.readInt32());
    }

    public writeSymbolModifier(modifier: SymbolModifier) {
        this.writeInt32(modifier.getModifier());
        this.writeInt32(modifier.getVisibility());
    }

    public readSymbolModifier(): SymbolModifier {
        return new SymbolModifier(this.readInt32(), this.readInt32());
    }

    public writeNamespaceName(namespace: NamespaceName) {
        this.writeString(namespace.parts.join('\\'));
    }

    public readNamespaceName(): NamespaceName {
        let namespace = new NamespaceName();

        namespace.parts = this.readString().split('\\');

        return namespace;
    }

    public getBuffer(): Buffer {
        return this.buffer.slice(0, this.length);
    }
}