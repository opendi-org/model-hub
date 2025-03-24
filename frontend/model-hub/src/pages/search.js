import React, { useState } from 'react';
import { Container, TextField, IconButton, List, ListItem, ListItemText, Box } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import AddIcon from '@mui/icons-material/Add';
import API_URL from '../config';
import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';


const SearchPage = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [results, setResults] = useState([]);

    const handleSearchChange = (event) => {
        setSearchTerm(event.target.value);
    };

    const handleSearch = () => {
        fetch(`${API_URL}/v0/models/search/${searchType}/${searchTerm}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                setResults(data)
            }).catch(error => console.error('There was an error fetching search results:', error));
    };

    const [searchType, setSearchType] = React.useState('model');

    const handleChange = (event) => {
        setSearchType(event.target.value);
    };

    return (
        <Container>
            <Box display="flex" justifyContent="center" alignItems="center" mt={2}>
                <TextField
                    variant="outlined"
                    placeholder="Search"
                    value={searchTerm}
                    onChange={handleSearchChange}
                    InputProps={{
                        startAdornment: (
                            <IconButton onClick={handleSearch}>
                                <SearchIcon />
                            </IconButton>
                        ),
                    }}
                />
                <FormControl>
                <InputLabel>Filter</InputLabel>
                <Select
                    labelId="select-search"
                    id="select-search"
                    value={searchType}
                    label="Filter"
                    onChange={handleChange}
                >
                    <MenuItem value={"model"}>Model Name</MenuItem>
                    <MenuItem value={"user"}>Creator Name</MenuItem>
                </Select>
                </FormControl>
                            </Box>
            <List>
                {results.map((result, index) => (
                    <ListItem key={index} divider>
                        <ListItemText
                            primary={result.name}
                            secondary={`Tags: ${result.tags}\n${result.description}`}
                        />
                    </ListItem>
                ))}
            </List>
        </Container>
    );
};

export default SearchPage