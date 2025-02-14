import { Typography} from '@mui/material';
import Grid from '@mui/material/Grid2';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import { NavLink } from "react-router-dom";

const ModelCard = ({id}) => {
    return (
    <Grid item xs={4}>
        <Card sx={{ minWidth: 275 }}>
            <CardContent>
                <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                    Lorem Ipsum1
                </Typography>
                <Typography variant="body2">
                    Lorem Ipsum
                    <br />
                    Lorem Ipsum
                    <br />
                    Lorem Ipsum
                </Typography>
                <Button 
                    variant="contained" 
                    color="primary" 
                    component={NavLink} to="/model/${id}"
                >
                    View
                </Button>
            </CardContent>
        </Card>
    </Grid>
)};

export default ModelCard;