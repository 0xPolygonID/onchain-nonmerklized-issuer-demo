import React, { useEffect } from 'react';
import { FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import { SelectChangeEvent } from '@mui/material/Select';

interface ListProps {
    datalist: string[];
    label: string;
    callback?: (issuer: string) => void;
}

const Selecter: React.FC<ListProps> = ({ datalist, label, callback }) => {
    const [selectedValue, setSelectedValue] = React.useState('');
    
    useEffect(() => {
        if (datalist.length > 0) {
            setSelectedValue(datalist[0]);
            callback && callback(datalist[0]);
        }
    }, [datalist, callback]);

    const handleChange = (event: SelectChangeEvent) => {
        const selectedValue = event.target.value as string;
        setSelectedValue(selectedValue);
        callback && callback(selectedValue);
    };
    return (
        <FormControl variant="standard" sx={{maxWidth: '100%', minWidth: '100%'}}>
            <InputLabel id="select-label">{label}</InputLabel>
            <Select
                labelId="select-label"
                id="select"
                value={selectedValue}
                onChange={handleChange}
                autoWidth
            >
                {datalist.map((v) => (
                    <MenuItem key={v} value={v}>
                        {v}
                    </MenuItem>
                ))}
            </Select>
        </FormControl>
    );
};

export default Selecter;
