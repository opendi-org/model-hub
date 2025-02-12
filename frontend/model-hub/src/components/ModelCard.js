import { Typography} from '@mui/material';
import Grid from '@mui/material/Grid2';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';

const ModelCard = () => {
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
            </CardContent>
        </Card>
    </Grid>
)};

export default ModelCard;