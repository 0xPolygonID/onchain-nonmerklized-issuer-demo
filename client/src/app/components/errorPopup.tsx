import React from 'react';
import { Alert, Box, Collapse, IconButton } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';

interface ErrorPopupProps {
    error: string;
}

const ErrorPopup: React.FC<ErrorPopupProps> = ({ error }) => {
    const [open, setOpen] = React.useState(true);

    return (
    <Box sx={{ position: 'fixed', top: '5%', left: '50%', transform: 'translate(-50%, -50%)', zIndex: 9999 }}>
        <Collapse in={open}>
            <Alert
                action={
                    <IconButton
                        aria-label="close"
                        color="inherit"
                        size="small"
                        onClick={() => {
                            setOpen(false);
                        }}
                    >
                        <CloseIcon fontSize="inherit" />
                    </IconButton>
                }
                severity="error"
            >
                {error}
            </Alert>
        </Collapse>
    </Box>
    );
};

export default ErrorPopup;
