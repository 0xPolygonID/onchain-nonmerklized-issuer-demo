import { FunctionComponent } from "react";
import QRCode from 'qrcode.react';
import { Container, Box, Button} from '@mui/material';

interface CodeProps {
    value: string;
}

const Code: FunctionComponent<CodeProps> = (props) => {
  const { value } = props;
  return (
    <Container>
      <Box sx={{ width: [400] }}>
        <QRCode level={"L"} size={400} value={value} />
      </Box>
      <Button onClick={() => dispachEvent(value)} 
        variant="contained" size="large" sx={{
          width: "100%",
          marginTop: "15px"
        }}> Polygon ID </Button>
    </Container>
  );
};

const dispachEvent = async (value: string) => {
  console.log('data to ext:', value);
  const msg = btoa(value);
  const hrefValue = `iden3comm://?i_m=${msg}`;
  console.log('link to ext:', hrefValue);

  const _authEvent = new CustomEvent('authEvent', { detail: hrefValue });
  document.dispatchEvent(_authEvent);
}

export default Code;
