import contractABI from './abi.json';
import Web3 from 'web3';
import {Id} from '@iden3/js-iden3-core';

export const isOnchainIssuerInterfaceImplemented = async (contractAddress: string): Promise<boolean> => {
    const web3 = new Web3(window.ethereum);
    const contract = new web3.eth.Contract(contractABI, contractAddress);
    const isImplemented = await contract.methods["supportsInterface"]("0x58874949").call();
    return isImplemented;
}

export const issueCredential = async (contractAddress: string, userId: Id) => {
    const web3 = new Web3(window.ethereum);
    const accounts = await web3.eth.getAccounts();
    const from = accounts[0];
    const onchainNonMerklizedIssuer = new web3.eth.Contract(contractABI, contractAddress);

    const estimatedGas = await onchainNonMerklizedIssuer.methods["issueCredential"](userId.bigInt()).estimateGas({ from });
    const gasLimit =estimatedGas + (estimatedGas * BigInt(15)) / BigInt(100);

    await onchainNonMerklizedIssuer.methods["issueCredential"](userId.bigInt()).send({ from, gas: gasLimit.toString() });
};

export const userCredentialIds = async (contractAddress: string, userId: Id): Promise<Array<string>> => {
    const web3 = new Web3(window.ethereum);
    const contract = new web3.eth.Contract(contractABI, contractAddress);
    const result = await contract.methods["getUserCredentialIds"](userId.bigInt()).call();
    if (!Array.isArray(result)) {
        throw new Error('Invalid result');
    }
    return result;
}

export const getCredential = async (contractAddress: string, userId: Id, credentialId: string): Promise<string> => {
    const web3 = new Web3(window.ethereum);
    const functionAbi = contractABI.find(func => func.name === "getCredential" && func.type === "function");
    if (!functionAbi) {
        throw new Error('Function ABI not found');
    }
    const data = web3.eth.abi.encodeFunctionCall(functionAbi, [userId.bigInt(), credentialId]);
    const transactionObject = {
        to: contractAddress.startsWith('0x') ? contractAddress : '0x' + contractAddress,
        data: data
    };    
    const resultHex = await web3.eth.call(transactionObject);
    console.log('Raw hex result', resultHex);
    return resultHex;
}

export const getOnchainIssuerVersion = async (contractAddress: string): Promise<string> => {
    const web3 = new Web3(window.ethereum);
    const contract = new web3.eth.Contract(contractABI, contractAddress);
    const result = await contract.methods["getCredentialAdapterVersion"]().call();
    return result;
}