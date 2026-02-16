import React, { useState, useEffect, useCallback } from 'react';
import { Container, TextField, IconButton, Box } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import API_URL from '../config';
import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';
import ModelMinicard from '../components/ModelMinicard'
import { useSearchParams } from "react-router-dom";


const SearchPage = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const [searchTerm, setSearchTerm] = useState(searchParams.get('term') ?? '');
    const [results, setResults] = useState([]);
    const [searchType, setSearchType] = useState(searchParams.get('type') ?? 'model');

    const handleSearchChange = (event) => {
        setSearchTerm(event.target.value);
    };

    const handleSearch = useCallback(() => {
        if (!searchTerm.trim()) {
            setResults([]);
            return;
        }

        // Update URL params when searching
        const newParams = new URLSearchParams();
        if (searchTerm.trim()) {
            newParams.set('term', searchTerm.trim());
        }
        if (searchType) {
            newParams.set('type', searchType);
        }
        setSearchParams(newParams);

        // URL encode the search term for the API call
        const encodedTerm = encodeURIComponent(searchTerm.trim());
        fetch(`${API_URL}/v0/models/search/${searchType}/${encodedTerm}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                setResults(data || []);
            })
            .catch(error => {
                console.error('There was an error fetching search results:', error);
                setResults([]);
            });
    }, [searchTerm, searchType, setSearchParams]);

    // Run search when URL params change (e.g., from navbar search)
    useEffect(() => {
        const termFromUrl = searchParams.get('term') ?? '';
        const typeFromUrl = searchParams.get('type') ?? 'model';
        
        // Update local state to match URL params
        setSearchTerm(termFromUrl);
        setSearchType(typeFromUrl);
        
        // Trigger search if there's a term
        if (termFromUrl.trim()) {
            const encodedTerm = encodeURIComponent(termFromUrl.trim());
            fetch(`${API_URL}/v0/models/search/${typeFromUrl}/${encodedTerm}`)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(data => {
                    setResults(data || []);
                })
                .catch(error => {
                    console.error('There was an error fetching search results:', error);
                    setResults([]);
                });
        } else {
            setResults([]);
        }
    }, [searchParams]);

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
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                            handleSearch();
                        }
                    }}
                />
                <FormControl sx={{ ml: 1, minWidth: 120 }}>
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
            <Box display="flex" flexDirection="column" alignItems="center" mt={2}>
                {results.map((result) => {
                    // Handle both uuid and UUID casing (use whichever exists)
                    const uuid = result.meta?.uuid || result.meta?.UUID;
                    if (!uuid) {
                        console.warn('Result missing UUID:', result);
                        return null;
                    }
                    return (
                        <React.Fragment key={uuid}>
                            <ModelMinicard 
                                name={result.meta?.name} 
                                id={uuid} 
                                author={result.meta?.creator?.username || 'Unknown'} 
                                summary={result.meta?.summary || ''} 
                                version={result.meta?.version} 
                                updatedDate={result.meta?.updatedDate}
                            />
                            <Box mb={2} /> 
                        </React.Fragment>
                    );
                })}
            </Box>
        </Container>
    );
};

export default SearchPage