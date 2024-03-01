import React, { Dispatch } from 'react';

interface SelectedIssuerContextProps {
    selectedIssuerContext: string;
    setSelectedIssuerContext: Dispatch<React.SetStateAction<string>>;
}

const SelectedIssuerContext = React.createContext<SelectedIssuerContextProps>({
    selectedIssuerContext: '',
    setSelectedIssuerContext: () => {},
});


export default SelectedIssuerContext;
