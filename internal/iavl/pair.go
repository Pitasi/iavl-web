package iavl

import (
	"errors"
	"fmt"

	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
)

type DBPair struct {
	Left  db.DB
	Right db.DB
}

func OpenDBPair(left, right string) (*DBPair, error) {
	leftDB, err := OpenDB(left)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", left, err)
	}

	rightDB, err := OpenDB(right)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", right, err)
	}

	return &DBPair{
		Left:  leftDB,
		Right: rightDB,
	}, nil
}

func (p *DBPair) Stats() (S, S) {
	left := Stats(p.Left)
	right := Stats(p.Right)
	return left, right
}

func (p *DBPair) Close() error {
	errL := p.Left.Close()
	errR := p.Right.Close()
	return errors.Join(errL, errR)
}

func (p *DBPair) LoadTrees(version int, prefix []byte) (*IAVLPair, error) {
	treeLeft, err := ReadTree(p.Left, version, prefix)
	if err != nil {
		return nil, err
	}
	treeRight, err := ReadTree(p.Right, version, prefix)
	if err != nil {
		return nil, err
	}
	return &IAVLPair{
		Left:  treeLeft,
		Right: treeRight,
	}, nil
}

type IAVLPair struct {
	Left  *iavl.MutableTree
	Right *iavl.MutableTree
}

func (p *IAVLPair) LoadHigherCommonVersion() (int, error) {
	leftVersions := p.Left.AvailableVersions()
	rightVersions := p.Right.AvailableVersions()

	if len(leftVersions) == 0 {
		return 0, fmt.Errorf("no versions available in left tree")
	}

	if len(rightVersions) == 0 {
		return 0, fmt.Errorf("no versions available in right tree")
	}

	var commonV int
	for _, v := range leftVersions {
		for _, v2 := range rightVersions {
			if v == v2 && v > commonV {
				commonV = v
			}
		}
	}

	if commonV == 0 {
		return 0, fmt.Errorf("The two IAVL trees have no version in common.\nAvailable versions in the left db: %v\nAvailable versions in the right db: %v", leftVersions, rightVersions)
	}

	return commonV, nil
}

func (p *IAVLPair) Hash() ([]byte, []byte) {
	return p.Left.Hash(), p.Right.Hash()
}

func (p *IAVLPair) Close() error {
	errL := p.Left.Close()
	errR := p.Right.Close()
	return errors.Join(errL, errR)
}
