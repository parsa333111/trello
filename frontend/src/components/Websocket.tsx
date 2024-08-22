
let websocket: WebSocket | null = null;
type CallbackFunction = (data: any) => void;

export const connect = (callback: CallbackFunction) => {
    let accessToken = localStorage.getItem("accessToken");
    let url = `/api/ws/${accessToken}`;
    console.log('Attempting to connect to websocket:', url);

    console.log('Initializing websocket connection');
    websocket = new WebSocket(url);

    websocket.onopen = () => {
        console.log('Websocket connection established');
    };

    websocket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log('Received websocket message:', data);
        callback(data)
    };

    websocket.onerror = (error) => {
        console.error('Websocket error:', error);
    };

    websocket.onclose = () => {
        console.log('Websocket connection closed');
    };
};

const close = () => {
    if (websocket) {
        websocket.close();
    }
};

export default { connect, close }