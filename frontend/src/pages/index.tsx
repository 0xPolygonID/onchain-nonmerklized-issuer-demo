import Button from '@mui/material/Button';
import Box from '@mui/material/Grid';

const App = () => {
    return (
    <Box display="flex"
        justifyContent="center"
        alignItems="center"
        minHeight="100vh">
        <Button href="/signin" variant="contained" size="large">Sign Up</Button>
    </Box>
    )
}

export default App
