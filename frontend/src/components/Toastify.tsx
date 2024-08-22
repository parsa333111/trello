import axios, { AxiosError } from 'axios';
import toast from 'typescript-toastify';

export enum MessageType {
    Success = "success",
    Error = "error",
    Warning = "warning",
    Info = "info"
}

export const toastifyMessage = (message: string, messageType: MessageType) => {
    new toast({
        position: "top-center",
        toastMsg: message,
        autoCloseTime: 3000,
        canClose: true,
        showProgress: true,
        pauseOnHover: true,
        pauseOnFocusLoss: true,
        type: messageType,
        theme: "dark"
    });
};


export const handleErrorWihtToast = (error: AxiosError | Error, message: string) => {
    if (axios.isAxiosError(error)) {
        if (error.response) {
            console.error('Error Status:', error.response.status);
            console.error('Error Data:', error.response.data.message);
            toastifyMessage(message, MessageType.Error);
        } else if (error.request) {
            console.error('Error Request:', error.request);
            toastifyMessage("No connection.", MessageType.Error);
        } else {
            console.error('Error Message:', error.message);
            toastifyMessage("Request failed.", MessageType.Error);
        }
    } else {
        console.error('Error Non-Axios:', error);
        toastifyMessage("Request failed.", MessageType.Error);
    }
};