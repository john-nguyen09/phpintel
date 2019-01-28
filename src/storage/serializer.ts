import { TypeName } from "../type/name";
import { Location } from "../symbol/meta/location";
import { Range } from "../symbol/meta/range";
import { SymbolModifier } from "../symbol/meta/modifier";
import { TypeComposite } from "../type/composite";
import { NamespaceName } from "../symbol/name/namespaceName";

export class Serializer {
    public static readonly DEFAULT_SIZE = 512;

    private buffer: Buffer;
    private offset: number;
    private length: number;

    constructor(initialSize?: number) {
        if (initialSize === undefined) {
            initialSize = Serializer.DEFAULT_SIZE;
        }

        this.buffer = Buffer.alloc(initialSize);
        this.length = 0;
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

    public setBuffer(buffer: Buffer) {
        this.needs(buffer.length);
        buffer.copy(this.buffer, this.offset);
        this.offset += buffer.length;
        this.length += buffer.length;
    }

    public setString(str: string) {
        let strBuffer = Buffer.from(str, 'utf8');

        this.setInt32(strBuffer.length);
        this.setBuffer(strBuffer);
    }

    public setInt32(value: number) {
        this.needs(4);
        this.buffer.writeInt32BE(value, this.offset);
        this.offset += 4;
        this.length += 4;
    }

    public setBool(value: boolean) {
        this.needs(1);
        this.buffer.writeUInt8(value ? 1 : 0, this.offset);
        this.offset += 1;
        this.length += 1;
    }

    public setTypeComposite(types: TypeComposite) {
        this.setInt32(types.types.length);
        for (let type of types.types) {
            this.setTypeName(type);
        }
    }

    public setTypeName(name: TypeName | null) {
        if (name == null) {
            this.setBool(false);
        } else {
            this.setBool(true);
            this.setString(name.name);
        }
    }

    public setLocation(location: Location) {
        if (location.isEmpty) {
            this.setBool(false);
        } else {
            this.setBool(true);
            this.setString(location.uri);
            this.setRange(location.range);
        }
    }

    public setRange(range: Range) {
        this.setInt32(range.start);
        this.setInt32(range.end);
    }

    public setSymbolModifier(modifier: SymbolModifier) {
        this.setInt32(modifier.getModifier());
        this.setInt32(modifier.getVisibility());
    }

    public setNamespaceName(namespace: NamespaceName) {
        this.setString(namespace.parts.join('\\'));
    }

    public getBuffer(): Buffer {
        return this.buffer.slice(0, this.length);
    }
}

export class Deserializer {
    private buffer: Buffer;
    private offset: number;

    constructor(buffer: Buffer) {
        this.buffer = buffer;
        this.offset = 0;
    }

    public readBuffer(): Buffer {
        return this.buffer.slice(this.offset);
    }

    public readString(): string {
        let length = this.readInt32();
        let strBuffer = this.buffer.slice(this.offset, this.offset + length);
        this.offset += length;

        return strBuffer.toString('utf8');
    }

    public readInt32(): number {
        let value = this.buffer.readInt32BE(this.offset);
        this.offset += 4;

        return value;
    }

    public readBool(): boolean {
        let value = this.buffer.readUInt8(this.offset) == 1 ? true : false;
        this.offset += 1;

        return value;
    }

    public readTypeComposite(): TypeComposite {
        let numTypes = this.readInt32();
        let types = new TypeComposite();
        for (let i = 0; i < numTypes; i++) {
            types.push(this.readTypeName() || new TypeName(''));
        }

        return types;
    }

    public readTypeName(): TypeName | null {
        let hasTypeName = this.readBool();

        if (!hasTypeName) {
            return null;
        }

        return new TypeName(this.readString());
    }

    public readLocation(): Location {
        let hasLocation = this.readBool();

        if (!hasLocation) {
            return new Location();
        }

        return new Location(this.readString(), this.readRange());
    }

    public readRange(): Range {
        return new Range(this.readInt32(), this.readInt32());
    }

    public readSymbolModifier(): SymbolModifier {
        return new SymbolModifier(this.readInt32(), this.readInt32());
    }

    public readNamespaceName(): NamespaceName {
        let namespace = new NamespaceName();

        namespace.parts = this.readString().split('\\');

        return namespace;
    }
}
