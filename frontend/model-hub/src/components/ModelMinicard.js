//
// COPYRIGHT OpenDI
//

import { Typography} from '@mui/material';
import Grid from '@mui/material/Grid2';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import { NavLink } from "react-router-dom";

const ModelMinicard = ({id, name, author}) => {
    return (
    <Grid item xs={4}>
        <Card sx={{ minWidth: 275 }}>
            <CardContent>
                <Typography gutterBottom sx={{ color: 'text.primary', fontSize: 14 }}>
                    {name}
                </Typography>
                <Typography gutterBottom sx={{ color: 'text.seconday', fontSize: 12 }}>
                    {author}
                </Typography>
                <Typography variant="body2">
                    Summary
                </Typography>
                <Button 
                    variant="contained" 
                    color="primary" 
                    component={NavLink} to={"model/" + id}
                >
                    View
                </Button>
            </CardContent>
        </Card>
    </Grid>
)};

export default ModelMinicard;