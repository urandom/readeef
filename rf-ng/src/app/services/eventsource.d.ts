interface Callback {
    (event: DataEvent): void; 
}

interface DataEvent extends Event {
    data: string;
    type: string;
    lastEventId: string;
}

declare class EventSource {
    addEventListener(type: string, listener?: Callback, options?: boolean | AddEventListenerOptions): void;
    dispatchEvent(evt: Event): boolean;
    removeEventListener(type: string, listener?: Callback, options?: boolean | EventListenerOptions): void;
    readyState: number;
    onmessage: Callback;
    close();
    constructor(name: string);
}