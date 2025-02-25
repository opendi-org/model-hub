//
// COPYRIGHT OpenDI
//

import { CardActions, Typography} from '@mui/material';
import Grid from '@mui/material/Grid2';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import { NavLink } from "react-router-dom";

const ModelMinicard = ({id, name, author}) => {
    return (
    <Grid xs={4}>
        <Card sx={{ minWidth: 275 }}>
            <CardContent>
                <Typography gutterBottom variant="h6" sx={{ color: 'text.primary'}}>
                    {name}
                </Typography>
                <Typography gutterBottom variant="body2" sx={{ color: 'text.secondary' }}>
                    {author}
                </Typography>
                <Typography variant="body2">
                    Summary
                </Typography>
            </CardContent>
            <CardActions>
            <Button 
                    variant="contained" 
                    color="primary" 
                    component={NavLink} to={"model/" + id}
                >
                    View
                </Button>
            </CardActions>
        </Card>
    </Grid>
)};

export default ModelMinicard;