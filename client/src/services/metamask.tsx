import Web3 from 'web3';
import {fromWei} from 'web3-utils';

declare global {
    interface Window {
        ethereum?: any;
    }
}

interface MetamaskError {
    message: string;
    code: number;
}

interface MetamaskWalletResponse {
    address: string;
    balance: string;
}

export async function selectMetamaskWallet(): Promise<MetamaskWalletResponse> {
    try {
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        if (!accounts || !accounts.length) {
            throw new Error('No accounts found');
        }

        const web3 = new Web3(window.ethereum);
        const balance = await web3.eth.getBalance(accounts[0]);
        const balanceInEther = web3.utils.fromWei(balance, 'ether');

        return {
            address: accounts[0],
            balance: balanceInEther,
        };
    } catch (error) {
        const metamaskError = error as MetamaskError;
        // user rejected request
        if (metamaskError.code === 4001) {
            throw new Error('User rejected request');
        }
        throw error;
    }
}
